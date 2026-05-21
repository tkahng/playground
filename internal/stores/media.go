package stores

import (
	"context"
	"log/slog"
	"slices"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
)

// ── Media store ───────────────────────────────────────────────────────────────

type MediaStoreInterface interface {
	CreateMedia(ctx context.Context, media *models.Medium) (*models.Medium, error)
	FindMediaByID(ctx context.Context, mediaId uuid.UUID) (*models.Medium, error)
	FindMediaByIDs(ctx context.Context, mediaIds []uuid.UUID) ([]*models.Medium, error)
	UpdateMedia(ctx context.Context, media *models.Medium) (*models.Medium, error)
	DeleteMedia(ctx context.Context, mediaId uuid.UUID) error
	FindMedia(ctx context.Context, filter *MediaListFilter) ([]*models.Medium, error)
	CountMedia(ctx context.Context, filter *MediaListFilter) (int64, error)
}

type DbMediaStore struct {
	dbx database.Dbx
}

func NewMediaStore(dbx database.Dbx) *DbMediaStore {
	return &DbMediaStore{dbx: dbx}
}

func (s *DbMediaStore) WithTx(dbx database.Dbx) *DbMediaStore {
	return &DbMediaStore{dbx: dbx}
}

func (s *DbMediaStore) CreateMedia(ctx context.Context, media *models.Medium) (*models.Medium, error) {
	return repository.Media.PostOne(ctx, s.dbx, media)
}

func (s *DbMediaStore) UpdateMedia(ctx context.Context, media *models.Medium) (*models.Medium, error) {
	return repository.Media.PutOne(ctx, s.dbx, media)
}

func (s *DbMediaStore) FindMediaByID(ctx context.Context, mediaId uuid.UUID) (*models.Medium, error) {
	return repository.Media.GetOne(ctx, s.dbx, &map[string]any{
		"id": map[string]any{"_eq": mediaId},
	})
}

func (s *DbMediaStore) DeleteMedia(ctx context.Context, mediaId uuid.UUID) error {
	_, err := repository.Media.Delete(ctx, s.dbx, &map[string]any{
		"id": map[string]any{"_eq": mediaId},
	})
	return err
}

func (s *DbMediaStore) FindMediaByIDs(ctx context.Context, mediaIds []uuid.UUID) ([]*models.Medium, error) {
	if len(mediaIds) == 0 {
		return nil, nil
	}
	return repository.Media.Get(ctx, s.dbx, &map[string]any{
		"id": map[string]any{"_in": mediaIds},
	}, nil, nil, nil)
}

type MediaListFilter struct {
	PaginatedInput
	SortParams
	Q       string      `query:"q,omitempty" required:"false"`
	UserIds []uuid.UUID `query:"userId,omitempty" format:"uuid" required:"false"`
}

func (s *DbMediaStore) FindMedia(ctx context.Context, filter *MediaListFilter) ([]*models.Medium, error) {
	if filter == nil {
		filter = &MediaListFilter{}
	}
	where := s.filter(filter)
	orderBy := s.sort(filter)
	limit, offset := pagination(filter)
	return repository.Media.Get(ctx, s.dbx, where, orderBy, &limit, &offset)
}

func (s *DbMediaStore) CountMedia(ctx context.Context, filter *MediaListFilter) (int64, error) {
	return repository.Media.Count(ctx, s.dbx, s.filter(filter))
}

func (s *DbMediaStore) filter(filter *MediaListFilter) *map[string]any {
	if filter == nil {
		return nil
	}
	where := make(map[string]any)
	if len(filter.UserIds) > 0 {
		where["user_id"] = map[string]any{"_in": filter.UserIds}
	}
	if filter.Q != "" {
		where["_or"] = []map[string]any{
			{"original_filename": map[string]any{"_ilike": "%" + filter.Q + "%"}},
			{"alt_text": map[string]any{"_ilike": "%" + filter.Q + "%"}},
			{"storage_key": map[string]any{"_ilike": "%" + filter.Q + "%"}},
		}
	}
	return &where
}

func (s *DbMediaStore) sort(filter Sortable) *map[string]string {
	if filter == nil {
		return nil
	}
	sortBy, sortOrder := filter.Sort()
	if sortBy != "" && slices.Contains(repository.MediaBuilder.FieldNames(), sortBy) {
		return &map[string]string{sortBy: sortOrder}
	}
	slog.Info("sort by field not found", "sortBy", sortBy)
	return nil
}

// ── Media decorator ───────────────────────────────────────────────────────────

type MediaStoreDecorator struct {
	Delegate            MediaStoreInterface
	CountMediaFunc      func(ctx context.Context, filter *MediaListFilter) (int64, error)
	CreateMediaFunc     func(ctx context.Context, media *models.Medium) (*models.Medium, error)
	FindMediaFunc       func(ctx context.Context, filter *MediaListFilter) ([]*models.Medium, error)
	FindMediaByIDFunc   func(ctx context.Context, mediaId uuid.UUID) (*models.Medium, error)
	FindMediaByIDsFunc  func(ctx context.Context, mediaIds []uuid.UUID) ([]*models.Medium, error)
	UpdateMediaFunc     func(ctx context.Context, media *models.Medium) (*models.Medium, error)
	DeleteMediaFunc     func(ctx context.Context, mediaId uuid.UUID) error
}

func (m *MediaStoreDecorator) CountMedia(ctx context.Context, filter *MediaListFilter) (int64, error) {
	if m.CountMediaFunc != nil {
		return m.CountMediaFunc(ctx, filter)
	}
	return m.Delegate.CountMedia(ctx, filter)
}

func (m *MediaStoreDecorator) CreateMedia(ctx context.Context, media *models.Medium) (*models.Medium, error) {
	if m.CreateMediaFunc != nil {
		return m.CreateMediaFunc(ctx, media)
	}
	return m.Delegate.CreateMedia(ctx, media)
}

func (m *MediaStoreDecorator) FindMedia(ctx context.Context, filter *MediaListFilter) ([]*models.Medium, error) {
	if m.FindMediaFunc != nil {
		return m.FindMediaFunc(ctx, filter)
	}
	return m.Delegate.FindMedia(ctx, filter)
}

func (m *MediaStoreDecorator) FindMediaByID(ctx context.Context, mediaId uuid.UUID) (*models.Medium, error) {
	if m.FindMediaByIDFunc != nil {
		return m.FindMediaByIDFunc(ctx, mediaId)
	}
	return m.Delegate.FindMediaByID(ctx, mediaId)
}

func (m *MediaStoreDecorator) FindMediaByIDs(ctx context.Context, mediaIds []uuid.UUID) ([]*models.Medium, error) {
	if m.FindMediaByIDsFunc != nil {
		return m.FindMediaByIDsFunc(ctx, mediaIds)
	}
	return m.Delegate.FindMediaByIDs(ctx, mediaIds)
}

func (m *MediaStoreDecorator) UpdateMedia(ctx context.Context, media *models.Medium) (*models.Medium, error) {
	if m.UpdateMediaFunc != nil {
		return m.UpdateMediaFunc(ctx, media)
	}
	return m.Delegate.UpdateMedia(ctx, media)
}

func (m *MediaStoreDecorator) DeleteMedia(ctx context.Context, mediaId uuid.UUID) error {
	if m.DeleteMediaFunc != nil {
		return m.DeleteMediaFunc(ctx, mediaId)
	}
	return m.Delegate.DeleteMedia(ctx, mediaId)
}

var _ MediaStoreInterface = (*MediaStoreDecorator)(nil)

func NewMediaStoreDecorator(dbx database.Dbx) *MediaStoreDecorator {
	return &MediaStoreDecorator{Delegate: NewMediaStore(dbx)}
}

// ── MediaAttachment store ─────────────────────────────────────────────────────

type MediaAttachmentStoreInterface interface {
	CreateAttachment(ctx context.Context, a *models.MediaAttachment) (*models.MediaAttachment, error)
	DeleteAttachment(ctx context.Context, entityType string, entityID uuid.UUID, slot string) error
	FindAttachments(ctx context.Context, entityType string, entityID uuid.UUID) ([]*models.MediaAttachment, error)
}

type DbMediaAttachmentStore struct {
	dbx database.Dbx
}

func NewMediaAttachmentStore(dbx database.Dbx) *DbMediaAttachmentStore {
	return &DbMediaAttachmentStore{dbx: dbx}
}

func (s *DbMediaAttachmentStore) WithTx(dbx database.Dbx) *DbMediaAttachmentStore {
	return &DbMediaAttachmentStore{dbx: dbx}
}

func (s *DbMediaAttachmentStore) CreateAttachment(ctx context.Context, a *models.MediaAttachment) (*models.MediaAttachment, error) {
	return repository.MediaAttachment.PostOne(ctx, s.dbx, a)
}

func (s *DbMediaAttachmentStore) DeleteAttachment(ctx context.Context, entityType string, entityID uuid.UUID, slot string) error {
	_, err := repository.MediaAttachment.Delete(ctx, s.dbx, &map[string]any{
		"entity_type": map[string]any{"_eq": entityType},
		"entity_id":   map[string]any{"_eq": entityID},
		"slot":        map[string]any{"_eq": slot},
	})
	return err
}

func (s *DbMediaAttachmentStore) FindAttachments(ctx context.Context, entityType string, entityID uuid.UUID) ([]*models.MediaAttachment, error) {
	return repository.MediaAttachment.Get(ctx, s.dbx, &map[string]any{
		"entity_type": map[string]any{"_eq": entityType},
		"entity_id":   map[string]any{"_eq": entityID},
	}, nil, nil, nil)
}

var _ MediaAttachmentStoreInterface = (*DbMediaAttachmentStore)(nil)
