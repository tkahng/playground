package stores

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/apierrors"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/tools/slug"
)

type BlogStoreInterface interface {
	CreatePost(ctx context.Context, input *CreateBlogPostDTO) (*models.BlogPost, error)
	UpdatePost(ctx context.Context, postID uuid.UUID, input *UpdateBlogPostDTO) (*models.BlogPost, error)
	PublishPost(ctx context.Context, postID uuid.UUID) (*models.BlogPost, error)
	UnpublishPost(ctx context.Context, postID uuid.UUID) (*models.BlogPost, error)
	ArchivePost(ctx context.Context, postID uuid.UUID) (*models.BlogPost, error)
	DeletePost(ctx context.Context, postID uuid.UUID) error
	FindPostByID(ctx context.Context, postID uuid.UUID) (*models.BlogPost, error)
	FindPostBySlug(ctx context.Context, postSlug string) (*models.BlogPost, error)
	ListPosts(ctx context.Context, filter *BlogPostFilter) ([]*models.BlogPost, error)
	CountPosts(ctx context.Context, filter *BlogPostFilter) (int64, error)
	IncrementViewCount(ctx context.Context, postID uuid.UUID) error
	CreateTag(ctx context.Context, input *CreateBlogTagDTO) (*models.BlogTag, error)
	ListTags(ctx context.Context) ([]*models.BlogTag, error)
	SetPostTags(ctx context.Context, postID uuid.UUID, tagIDs []uuid.UUID) error
	WithTx(dbx database.Dbx) *DbBlogStore
}

type DbBlogStore struct {
	db database.Dbx
}

var _ BlogStoreInterface = (*DbBlogStore)(nil)

func NewDbBlogStore(db database.Dbx) *DbBlogStore {
	return &DbBlogStore{db: db}
}

func (s *DbBlogStore) WithTx(dbx database.Dbx) *DbBlogStore {
	return &DbBlogStore{db: dbx}
}

type BlogPostFilter struct {
	PaginatedInput
	SortParams
	Status   []models.BlogPostStatus `query:"status,omitempty" required:"false"`
	AuthorID *uuid.UUID              `query:"author_id,omitempty" required:"false" format:"uuid"`
	Q        string                  `query:"q,omitempty" required:"false"`
}

type CreateBlogPostDTO struct {
	Title            string                   `json:"title" required:"true" minLength:"1"`
	Content          string                   `json:"content" required:"false"`
	ContentFormat    models.BlogContentFormat `json:"content_format" required:"false" enum:"tiptap,markdown"`
	AuthorID         uuid.UUID                `json:"author_id" required:"true" format:"uuid"`
	FeaturedImageKey *string                  `json:"featured_image_key,omitempty" required:"false"`
	SeoTitle         *string                  `json:"seo_title,omitempty" required:"false"`
	SeoDescription   *string                  `json:"seo_description,omitempty" required:"false"`
	TagIDs           []uuid.UUID              `json:"tag_ids,omitempty" required:"false" format:"uuid"`
}

type UpdateBlogPostDTO struct {
	Title            *string                   `json:"title,omitempty" required:"false" minLength:"1"`
	Content          *string                   `json:"content,omitempty" required:"false"`
	ContentFormat    *models.BlogContentFormat `json:"content_format,omitempty" required:"false" enum:"tiptap,markdown"`
	FeaturedImageKey *string                   `json:"featured_image_key,omitempty" required:"false"`
	SeoTitle         *string                   `json:"seo_title,omitempty" required:"false"`
	SeoDescription   *string                   `json:"seo_description,omitempty" required:"false"`
	TagIDs           []uuid.UUID               `json:"tag_ids,omitempty" required:"false" format:"uuid"`
}

type CreateBlogTagDTO struct {
	Name string `json:"name" required:"true" minLength:"1"`
}

func (s *DbBlogStore) buildPostWhere(filter *BlogPostFilter) *map[string]any {
	where := map[string]any{}
	if filter == nil {
		return &where
	}
	if len(filter.Status) > 0 {
		where["status"] = map[string]any{"_in": filter.Status}
	}
	if filter.AuthorID != nil {
		where["author_id"] = map[string]any{"_eq": *filter.AuthorID}
	}
	return &where
}

func (s *DbBlogStore) buildPostOrder(filter *BlogPostFilter) *map[string]string {
	if filter == nil {
		return &map[string]string{"published_at": "desc"}
	}
	sortBy, sortOrder := filter.Sort()
	if sortBy == "" {
		sortBy = "published_at"
	}
	if sortOrder == "" {
		sortOrder = "desc"
	}
	return &map[string]string{sortBy: sortOrder}
}

func (s *DbBlogStore) uniqueSlug(ctx context.Context, base string) (string, error) {
	candidate := slug.NewSlug(base)
	if candidate == "" {
		candidate = "post"
	}
	for i := 0; i < 100; i++ {
		attempt := candidate
		if i > 0 {
			attempt = fmt.Sprintf("%s-%d", candidate, i+1)
		}
		_, err := repository.BlogPost.GetOne(ctx, s.db, &map[string]any{
			"slug": map[string]any{"_eq": attempt},
		})
		if err != nil {
			// not found → slug is free
			return attempt, nil
		}
	}
	return "", apierrors.Conflict("could not generate unique slug")
}

func estimateReadingTime(content string) int {
	words := 0
	inWord := false
	for _, r := range content {
		if unicode.IsSpace(r) || unicode.IsPunct(r) {
			inWord = false
		} else if !inWord {
			inWord = true
			words++
		}
	}
	if words == 0 {
		return 1
	}
	minutes := (words + 199) / 200
	if minutes < 1 {
		return 1
	}
	return minutes
}

func (s *DbBlogStore) CreatePost(ctx context.Context, input *CreateBlogPostDTO) (*models.BlogPost, error) {
	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" {
		return nil, apierrors.BadRequest("title is required")
	}
	if input.ContentFormat == "" {
		input.ContentFormat = models.BlogContentFormatTiptap
	}

	postSlug, err := s.uniqueSlug(ctx, input.Title)
	if err != nil {
		return nil, err
	}

	rt := estimateReadingTime(input.Content)
	post := &models.BlogPost{
		Slug:               postSlug,
		Title:              input.Title,
		Content:            input.Content,
		ContentFormat:      input.ContentFormat,
		Status:             models.BlogPostStatusDraft,
		AuthorID:           input.AuthorID,
		FeaturedImageKey:   input.FeaturedImageKey,
		SeoTitle:           input.SeoTitle,
		SeoDescription:     input.SeoDescription,
		ReadingTimeMinutes: &rt,
	}

	created, err := repository.BlogPost.PostOne(ctx, s.db, post)
	if err != nil {
		return nil, err
	}

	if len(input.TagIDs) > 0 {
		if err := s.SetPostTags(ctx, created.ID, input.TagIDs); err != nil {
			return nil, err
		}
	}

	return created, nil
}

func (s *DbBlogStore) UpdatePost(ctx context.Context, postID uuid.UUID, input *UpdateBlogPostDTO) (*models.BlogPost, error) {
	post, err := repository.BlogPost.GetOne(ctx, s.db, &map[string]any{
		"id": map[string]any{"_eq": postID},
	})
	if err != nil {
		return nil, apierrors.NotFound("post not found")
	}

	if input.Title != nil {
		t := strings.TrimSpace(*input.Title)
		if t == "" {
			return nil, apierrors.BadRequest("title cannot be empty")
		}
		post.Title = t
	}
	if input.Content != nil {
		post.Content = *input.Content
		rt := estimateReadingTime(post.Content)
		post.ReadingTimeMinutes = &rt
	}
	if input.ContentFormat != nil {
		post.ContentFormat = *input.ContentFormat
	}
	if input.FeaturedImageKey != nil {
		post.FeaturedImageKey = input.FeaturedImageKey
	}
	if input.SeoTitle != nil {
		post.SeoTitle = input.SeoTitle
	}
	if input.SeoDescription != nil {
		post.SeoDescription = input.SeoDescription
	}

	updated, err := repository.BlogPost.PutOne(ctx, s.db, post)
	if err != nil {
		return nil, err
	}

	if input.TagIDs != nil {
		if err := s.SetPostTags(ctx, postID, input.TagIDs); err != nil {
			return nil, err
		}
	}

	return updated, nil
}

func (s *DbBlogStore) PublishPost(ctx context.Context, postID uuid.UUID) (*models.BlogPost, error) {
	post, err := repository.BlogPost.GetOne(ctx, s.db, &map[string]any{
		"id": map[string]any{"_eq": postID},
	})
	if err != nil {
		return nil, apierrors.NotFound("post not found")
	}
	now := time.Now()
	post.Status = models.BlogPostStatusPublished
	post.PublishedAt = &now
	return repository.BlogPost.PutOne(ctx, s.db, post)
}

func (s *DbBlogStore) UnpublishPost(ctx context.Context, postID uuid.UUID) (*models.BlogPost, error) {
	post, err := repository.BlogPost.GetOne(ctx, s.db, &map[string]any{
		"id": map[string]any{"_eq": postID},
	})
	if err != nil {
		return nil, apierrors.NotFound("post not found")
	}
	post.Status = models.BlogPostStatusDraft
	post.PublishedAt = nil
	return repository.BlogPost.PutOne(ctx, s.db, post)
}

func (s *DbBlogStore) ArchivePost(ctx context.Context, postID uuid.UUID) (*models.BlogPost, error) {
	post, err := repository.BlogPost.GetOne(ctx, s.db, &map[string]any{
		"id": map[string]any{"_eq": postID},
	})
	if err != nil {
		return nil, apierrors.NotFound("post not found")
	}
	post.Status = models.BlogPostStatusArchived
	return repository.BlogPost.PutOne(ctx, s.db, post)
}

func (s *DbBlogStore) DeletePost(ctx context.Context, postID uuid.UUID) error {
	_, err := repository.BlogPost.Delete(ctx, s.db, &map[string]any{
		"id": map[string]any{"_eq": postID},
	})
	return err
}

func (s *DbBlogStore) FindPostByID(ctx context.Context, postID uuid.UUID) (*models.BlogPost, error) {
	post, err := repository.BlogPost.GetOne(ctx, s.db, &map[string]any{
		"id": map[string]any{"_eq": postID},
	})
	if err != nil {
		return nil, apierrors.NotFound("post not found")
	}
	return post, nil
}

func (s *DbBlogStore) FindPostBySlug(ctx context.Context, postSlug string) (*models.BlogPost, error) {
	post, err := repository.BlogPost.GetOne(ctx, s.db, &map[string]any{
		"slug": map[string]any{"_eq": postSlug},
	})
	if err != nil {
		return nil, apierrors.NotFound("post not found")
	}
	return post, nil
}

func (s *DbBlogStore) ListPosts(ctx context.Context, filter *BlogPostFilter) ([]*models.BlogPost, error) {
	where := s.buildPostWhere(filter)
	order := s.buildPostOrder(filter)
	limit, offset := pagination(filter)
	return repository.BlogPost.Get(ctx, s.db, where, order, &limit, &offset)
}

func (s *DbBlogStore) CountPosts(ctx context.Context, filter *BlogPostFilter) (int64, error) {
	where := s.buildPostWhere(filter)
	return repository.BlogPost.Count(ctx, s.db, where)
}

func (s *DbBlogStore) IncrementViewCount(ctx context.Context, postID uuid.UUID) error {
	_, err := database.Exec(ctx, s.db,
		`update blog.posts set view_count = view_count + 1 where id = $1`, postID)
	return err
}

func (s *DbBlogStore) CreateTag(ctx context.Context, input *CreateBlogTagDTO) (*models.BlogTag, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, apierrors.BadRequest("tag name is required")
	}
	tagSlug := slug.NewSlug(name)
	existing, err := repository.BlogTag.GetOne(ctx, s.db, &map[string]any{
		"slug": map[string]any{"_eq": tagSlug},
	})
	if err == nil {
		return existing, nil
	}
	return repository.BlogTag.PostOne(ctx, s.db, &models.BlogTag{
		Name: name,
		Slug: tagSlug,
	})
}

func (s *DbBlogStore) ListTags(ctx context.Context) ([]*models.BlogTag, error) {
	order := map[string]string{"name": "asc"}
	where := map[string]any{}
	return repository.BlogTag.Get(ctx, s.db, &where, &order, nil, nil)
}

func (s *DbBlogStore) SetPostTags(ctx context.Context, postID uuid.UUID, tagIDs []uuid.UUID) error {
	_, err := database.Exec(ctx, s.db,
		`delete from blog.post_tags where post_id = $1`, postID)
	if err != nil {
		return err
	}
	if len(tagIDs) == 0 {
		return nil
	}
	args := []any{postID}
	placeholders := make([]string, len(tagIDs))
	for i, id := range tagIDs {
		args = append(args, id)
		placeholders[i] = fmt.Sprintf("($1, $%d)", i+2)
	}
	q := fmt.Sprintf(
		`insert into blog.post_tags (post_id, tag_id) values %s on conflict do nothing`,
		strings.Join(placeholders, ", "),
	)
	_, err = database.Exec(ctx, s.db, q, args...)
	return err
}
