package stores

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
)

type BlogStoreDecorator struct {
	Delegate BlogStoreInterface
}

var _ BlogStoreInterface = (*BlogStoreDecorator)(nil)

func NewBlogStoreDecorator(db database.Dbx) *BlogStoreDecorator {
	return &BlogStoreDecorator{Delegate: NewDbBlogStore(db)}
}

func (d *BlogStoreDecorator) CreatePost(ctx context.Context, input *CreateBlogPostDTO) (*models.BlogPost, error) {
	return d.Delegate.CreatePost(ctx, input)
}
func (d *BlogStoreDecorator) UpdatePost(ctx context.Context, postID uuid.UUID, input *UpdateBlogPostDTO) (*models.BlogPost, error) {
	return d.Delegate.UpdatePost(ctx, postID, input)
}
func (d *BlogStoreDecorator) PublishPost(ctx context.Context, postID uuid.UUID) (*models.BlogPost, error) {
	return d.Delegate.PublishPost(ctx, postID)
}
func (d *BlogStoreDecorator) UnpublishPost(ctx context.Context, postID uuid.UUID) (*models.BlogPost, error) {
	return d.Delegate.UnpublishPost(ctx, postID)
}
func (d *BlogStoreDecorator) ArchivePost(ctx context.Context, postID uuid.UUID) (*models.BlogPost, error) {
	return d.Delegate.ArchivePost(ctx, postID)
}
func (d *BlogStoreDecorator) DeletePost(ctx context.Context, postID uuid.UUID) error {
	return d.Delegate.DeletePost(ctx, postID)
}
func (d *BlogStoreDecorator) FindPostByID(ctx context.Context, postID uuid.UUID) (*models.BlogPost, error) {
	return d.Delegate.FindPostByID(ctx, postID)
}
func (d *BlogStoreDecorator) FindPostBySlug(ctx context.Context, postSlug string) (*models.BlogPost, error) {
	return d.Delegate.FindPostBySlug(ctx, postSlug)
}
func (d *BlogStoreDecorator) ListPosts(ctx context.Context, filter *BlogPostFilter) ([]*models.BlogPost, error) {
	return d.Delegate.ListPosts(ctx, filter)
}
func (d *BlogStoreDecorator) CountPosts(ctx context.Context, filter *BlogPostFilter) (int64, error) {
	return d.Delegate.CountPosts(ctx, filter)
}
func (d *BlogStoreDecorator) IncrementViewCount(ctx context.Context, postID uuid.UUID) error {
	return d.Delegate.IncrementViewCount(ctx, postID)
}
func (d *BlogStoreDecorator) CreateTag(ctx context.Context, input *CreateBlogTagDTO) (*models.BlogTag, error) {
	return d.Delegate.CreateTag(ctx, input)
}
func (d *BlogStoreDecorator) ListTags(ctx context.Context) ([]*models.BlogTag, error) {
	return d.Delegate.ListTags(ctx)
}
func (d *BlogStoreDecorator) SetPostTags(ctx context.Context, postID uuid.UUID, tagIDs []uuid.UUID) error {
	return d.Delegate.SetPostTags(ctx, postID, tagIDs)
}
func (d *BlogStoreDecorator) WithTx(dbx database.Dbx) *DbBlogStore {
	return d.Delegate.WithTx(dbx)
}
