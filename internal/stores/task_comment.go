package stores

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
)

type DbTaskCommentStoreInterface interface {
	CreateTaskComment(ctx context.Context, comment *models.TaskComment) (*models.TaskComment, error)
	FindTaskCommentByID(ctx context.Context, id uuid.UUID) (*models.TaskComment, error)
	ListTaskComments(ctx context.Context, taskID uuid.UUID) ([]*models.TaskComment, error)
	UpdateTaskComment(ctx context.Context, comment *models.TaskComment) (*models.TaskComment, error)
	DeleteTaskComment(ctx context.Context, id uuid.UUID) error
	WithTx(db database.Dbx) *DbTaskCommentStore
}

type DbTaskCommentStore struct {
	db database.Dbx
}

var _ DbTaskCommentStoreInterface = (*DbTaskCommentStore)(nil)

func NewDbTaskCommentStore(db database.Dbx) *DbTaskCommentStore {
	return &DbTaskCommentStore{db: db}
}

func (s *DbTaskCommentStore) WithTx(db database.Dbx) *DbTaskCommentStore {
	return &DbTaskCommentStore{db: db}
}

func (s *DbTaskCommentStore) CreateTaskComment(ctx context.Context, comment *models.TaskComment) (*models.TaskComment, error) {
	return repository.TaskComment.PostOne(ctx, s.db, comment)
}

func (s *DbTaskCommentStore) FindTaskCommentByID(ctx context.Context, id uuid.UUID) (*models.TaskComment, error) {
	return repository.TaskComment.GetOne(ctx, s.db, &map[string]any{
		"id": map[string]any{"_eq": id},
	})
}

func (s *DbTaskCommentStore) ListTaskComments(ctx context.Context, taskID uuid.UUID) ([]*models.TaskComment, error) {
	return repository.TaskComment.Get(ctx, s.db,
		&map[string]any{
			"task_id": map[string]any{"_eq": taskID},
		},
		&map[string]string{"created_at": "asc"},
		nil,
		nil,
	)
}

func (s *DbTaskCommentStore) UpdateTaskComment(ctx context.Context, comment *models.TaskComment) (*models.TaskComment, error) {
	return repository.TaskComment.PutOne(ctx, s.db, comment)
}

func (s *DbTaskCommentStore) DeleteTaskComment(ctx context.Context, id uuid.UUID) error {
	_, err := repository.TaskComment.Delete(ctx, s.db, &map[string]any{
		"id": map[string]any{"_eq": id},
	})
	return err
}
