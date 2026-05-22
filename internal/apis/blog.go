package apis

import (
	"context"
	"log/slog"
	"slices"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
)

// ── API types ──────────────────────────────────────────────────────────────

type BlogTag struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}

type BlogPost struct {
	ID                   uuid.UUID                `json:"id"`
	Slug                 string                   `json:"slug"`
	Title                string                   `json:"title"`
	Content              string                   `json:"content"`
	ContentFormat        models.BlogContentFormat `json:"content_format"`
	Status               models.BlogPostStatus    `json:"status"`
	AuthorID             uuid.UUID                `json:"author_id"`
	PublishedAt          *time.Time               `json:"published_at" nullable:"true"`
	FeaturedImageID      *uuid.UUID               `json:"featured_image_id" nullable:"true"`
	FeaturedImageURL     *string                  `json:"featured_image_url" nullable:"true"`
	SeoTitle             *string                  `json:"seo_title" nullable:"true"`
	SeoDescription       *string                  `json:"seo_description" nullable:"true"`
	ReadingTimeMinutes   *int                     `json:"reading_time_minutes" nullable:"true"`
	ViewCount            int64                    `json:"view_count"`
	Tags                 []*BlogTag               `json:"tags,omitempty"`
	CreatedAt            time.Time                `json:"created_at"`
	UpdatedAt            time.Time                `json:"updated_at"`
}

func fromModelBlogTag(t *models.BlogTag) *BlogTag {
	if t == nil {
		return nil
	}
	return &BlogTag{ID: t.ID, Name: t.Name, Slug: t.Slug, CreatedAt: t.CreatedAt}
}

// loadFeaturedMedia batch-fetches media records for the given posts' featured images.
// Returns a map of media_id → *models.Medium; missing IDs are simply absent from the map.
func (api *Api) loadFeaturedMedia(ctx context.Context, posts ...*models.BlogPost) map[uuid.UUID]*models.Medium {
	ids := make([]uuid.UUID, 0, len(posts))
	for _, p := range posts {
		if p.FeaturedImageID != nil {
			ids = append(ids, *p.FeaturedImageID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	medias, err := api.App().Adapter().Media().FindMediaByIDs(ctx, ids)
	if err != nil {
		slog.ErrorContext(ctx, "loadFeaturedMedia: FindMediaByIDs failed", "error", err)
		return nil
	}
	m := make(map[uuid.UUID]*models.Medium, len(medias))
	for _, med := range medias {
		m[med.ID] = med
	}
	return m
}

func (api *Api) fromModelBlogPost(p *models.BlogPost, media map[uuid.UUID]*models.Medium) *BlogPost {
	if p == nil {
		return nil
	}
	tags := make([]*BlogTag, len(p.Tags))
	for i, t := range p.Tags {
		tags[i] = fromModelBlogTag(t)
	}

	var featuredImageURL *string
	if p.FeaturedImageID != nil {
		if m, ok := media[*p.FeaturedImageID]; ok {
			if m.PublicURL != nil {
				featuredImageURL = m.PublicURL
			} else {
				u := api.App().Fs().PublicURL(m.StorageKey)
				featuredImageURL = &u
			}
		}
	}

	return &BlogPost{
		ID:                 p.ID,
		Slug:               p.Slug,
		Title:              p.Title,
		Content:            p.Content,
		ContentFormat:      p.ContentFormat,
		Status:             p.Status,
		AuthorID:           p.AuthorID,
		PublishedAt:        p.PublishedAt,
		FeaturedImageID:    p.FeaturedImageID,
		FeaturedImageURL:   featuredImageURL,
		SeoTitle:           p.SeoTitle,
		SeoDescription:     p.SeoDescription,
		ReadingTimeMinutes: p.ReadingTimeMinutes,
		ViewCount:          p.ViewCount,
		Tags:               tags,
		CreatedAt:          p.CreatedAt,
		UpdatedAt:          p.UpdatedAt,
	}
}

// ── List posts (public) ────────────────────────────────────────────────────

type BlogPostListInput struct {
	PaginatedInput
	SortParams
	Status   []string `query:"status,omitempty" required:"false" enum:"draft,published,archived"`
	AuthorID string   `query:"author_id,omitempty" required:"false" format:"uuid"`
	Q        string   `query:"q,omitempty" required:"false"`
}

func (api *Api) BlogPostList(ctx context.Context, input *BlogPostListInput) (*ApiPaginatedOutput[*BlogPost], error) {
	filter := &stores.BlogPostFilter{}
	filter.Page = input.Page
	filter.PerPage = input.PerPage
	filter.SortBy = input.SortBy
	filter.SortOrder = input.SortOrder
	filter.Q = input.Q

	userInfo := contextstore.GetContextUserInfo(ctx)
	isAdmin := userInfo != nil && slices.Contains(userInfo.Permissions, "superuser")
	if !isAdmin {
		filter.Status = []models.BlogPostStatus{models.BlogPostStatusPublished}
	} else {
		for _, s := range input.Status {
			filter.Status = append(filter.Status, models.BlogPostStatus(s))
		}
	}

	if input.AuthorID != "" {
		id, err := uuid.Parse(input.AuthorID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid author_id")
		}
		filter.AuthorID = &id
	}

	posts, err := api.App().Adapter().Blog().ListPosts(ctx, filter)
	if err != nil {
		return nil, err
	}
	count, err := api.App().Adapter().Blog().CountPosts(ctx, filter)
	if err != nil {
		return nil, err
	}

	media := api.loadFeaturedMedia(ctx, posts...)
	data := make([]*BlogPost, len(posts))
	for i, p := range posts {
		data[i] = api.fromModelBlogPost(p, media)
	}
	return &ApiPaginatedOutput[*BlogPost]{
		Body: ApiPaginatedResponse[*BlogPost]{
			Data: data,
			Meta: ApiGenerateMeta(&input.PaginatedInput, count),
		},
	}, nil
}

// ── Get post by slug (public) ──────────────────────────────────────────────

func (api *Api) BlogPostGet(ctx context.Context, input *struct {
	Slug string `path:"slug" required:"true"`
}) (*ApiSingleOutput[*BlogPost], error) {
	var post *models.BlogPost
	var err error
	// Accept a UUID (admin direct link by post ID) or a slug (public URL).
	if id, parseErr := uuid.Parse(input.Slug); parseErr == nil {
		post, err = api.App().Adapter().Blog().FindPostByID(ctx, id)
	} else {
		post, err = api.App().Adapter().Blog().FindPostBySlug(ctx, input.Slug)
	}
	if err != nil {
		return nil, err
	}

	userInfo := contextstore.GetContextUserInfo(ctx)
	isAdmin := userInfo != nil && slices.Contains(userInfo.Permissions, "superuser")
	if !isAdmin && post.Status != models.BlogPostStatusPublished {
		return nil, huma.Error404NotFound("post not found")
	}

	if post.Status == models.BlogPostStatusPublished {
		_ = api.App().Adapter().Blog().IncrementViewCount(ctx, post.ID)
	}

	media := api.loadFeaturedMedia(ctx, post)
	return &ApiSingleOutput[*BlogPost]{Body: ApiSingleResponse[*BlogPost]{Data: api.fromModelBlogPost(post, media)}}, nil
}

// requireAdmin returns the caller's UserInfo when they have the superuser
// permission, or a 401/403 error otherwise.  Call at the top of every admin
// handler as a defence-in-depth layer beneath the route middleware.
func requireAdmin(ctx context.Context) (*models.UserInfo, error) {
	userInfo := contextstore.GetContextUserInfo(ctx)
	if userInfo == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}
	if !slices.Contains(userInfo.Permissions, "superuser") {
		return nil, huma.Error403Forbidden("forbidden")
	}
	return userInfo, nil
}

// ── Create post (admin) ────────────────────────────────────────────────────

type BlogPostCreateInput struct {
	Body stores.CreateBlogPostDTO
}

func (api *Api) BlogPostCreate(ctx context.Context, input *BlogPostCreateInput) (*ApiSingleOutput[*BlogPost], error) {
	userInfo, err := requireAdmin(ctx)
	if err != nil {
		return nil, err
	}
	input.Body.AuthorID = userInfo.User.ID

	post, err := api.App().Adapter().Blog().CreatePost(ctx, &input.Body)
	if err != nil {
		return nil, err
	}
	media := api.loadFeaturedMedia(ctx, post)
	return &ApiSingleOutput[*BlogPost]{Body: ApiSingleResponse[*BlogPost]{Data: api.fromModelBlogPost(post, media)}}, nil
}

// ── Update post (admin) ────────────────────────────────────────────────────

type BlogPostUpdateInput struct {
	ID   string `path:"post-id" required:"true" format:"uuid"`
	Body stores.UpdateBlogPostDTO
}

func (api *Api) BlogPostUpdate(ctx context.Context, input *BlogPostUpdateInput) (*ApiSingleOutput[*BlogPost], error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	postID, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid post id")
	}
	post, err := api.App().Adapter().Blog().UpdatePost(ctx, postID, &input.Body)
	if err != nil {
		return nil, err
	}
	media := api.loadFeaturedMedia(ctx, post)
	return &ApiSingleOutput[*BlogPost]{Body: ApiSingleResponse[*BlogPost]{Data: api.fromModelBlogPost(post, media)}}, nil
}

// ── Publish post (admin) ───────────────────────────────────────────────────

func (api *Api) BlogPostPublish(ctx context.Context, input *struct {
	ID string `path:"post-id" required:"true" format:"uuid"`
}) (*ApiSingleOutput[*BlogPost], error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	postID, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid post id")
	}
	post, err := api.App().Adapter().Blog().PublishPost(ctx, postID)
	if err != nil {
		return nil, err
	}
	media := api.loadFeaturedMedia(ctx, post)
	return &ApiSingleOutput[*BlogPost]{Body: ApiSingleResponse[*BlogPost]{Data: api.fromModelBlogPost(post, media)}}, nil
}

// ── Unpublish post (admin) ─────────────────────────────────────────────────

func (api *Api) BlogPostUnpublish(ctx context.Context, input *struct {
	ID string `path:"post-id" required:"true" format:"uuid"`
}) (*ApiSingleOutput[*BlogPost], error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	postID, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid post id")
	}
	post, err := api.App().Adapter().Blog().UnpublishPost(ctx, postID)
	if err != nil {
		return nil, err
	}
	media := api.loadFeaturedMedia(ctx, post)
	return &ApiSingleOutput[*BlogPost]{Body: ApiSingleResponse[*BlogPost]{Data: api.fromModelBlogPost(post, media)}}, nil
}

// ── Archive post (admin) ───────────────────────────────────────────────────

func (api *Api) BlogPostArchive(ctx context.Context, input *struct {
	ID string `path:"post-id" required:"true" format:"uuid"`
}) (*ApiSingleOutput[*BlogPost], error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	postID, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid post id")
	}
	post, err := api.App().Adapter().Blog().ArchivePost(ctx, postID)
	if err != nil {
		return nil, err
	}
	media := api.loadFeaturedMedia(ctx, post)
	return &ApiSingleOutput[*BlogPost]{Body: ApiSingleResponse[*BlogPost]{Data: api.fromModelBlogPost(post, media)}}, nil
}

// ── Delete post (admin) ────────────────────────────────────────────────────

func (api *Api) BlogPostDelete(ctx context.Context, input *struct {
	ID string `path:"post-id" required:"true" format:"uuid"`
}) (*struct{}, error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	postID, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid post id")
	}
	if err := api.App().Adapter().Blog().DeletePost(ctx, postID); err != nil {
		return nil, err
	}
	return nil, nil
}

// ── Tags ───────────────────────────────────────────────────────────────────

func (api *Api) BlogTagList(ctx context.Context, _ *struct{}) (*ApiSingleOutput[[]*BlogTag], error) {
	tags, err := api.App().Adapter().Blog().ListTags(ctx)
	if err != nil {
		return nil, err
	}
	data := make([]*BlogTag, len(tags))
	for i, t := range tags {
		data[i] = fromModelBlogTag(t)
	}
	return &ApiSingleOutput[[]*BlogTag]{Body: ApiSingleResponse[[]*BlogTag]{Data: data}}, nil
}

type BlogTagCreateInput struct {
	Body stores.CreateBlogTagDTO
}

func (api *Api) BlogTagCreate(ctx context.Context, input *BlogTagCreateInput) (*ApiSingleOutput[*BlogTag], error) {
	if _, err := requireAdmin(ctx); err != nil {
		return nil, err
	}
	tag, err := api.App().Adapter().Blog().CreateTag(ctx, &input.Body)
	if err != nil {
		return nil, err
	}
	return &ApiSingleOutput[*BlogTag]{Body: ApiSingleResponse[*BlogTag]{Data: fromModelBlogTag(tag)}}, nil
}
