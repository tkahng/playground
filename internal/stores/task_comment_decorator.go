package stores

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database"
	"github.com/tkahng/playground/internal/models"
)

type TaskCommentStoreDecorator struct {
	Delegate DbTaskCommentStoreInterface
}

var _ DbTaskCommentStoreInterface = (*TaskCommentStoreDecorator)(nil)

func NewTaskCommentStoreDecorator(db database.Dbx) *TaskCommentStoreDecorator {
	return &TaskCommentStoreDecorator{Delegate: NewDbTaskCommentStore(db)}
}

func (d *TaskCommentStoreDecorator) CreateTaskComment(ctx context.Context, comment *models.TaskComment) (*models.TaskComment, error) {
	return d.Delegate.CreateTaskComment(ctx, comment)
}
func (d *TaskCommentStoreDecorator) FindTaskCommentByID(ctx context.Context, id uuid.UUID) (*models.TaskComment, error) {
	return d.Delegate.FindTaskCommentByID(ctx, id)
}
func (d *TaskCommentStoreDecorator) ListTaskComments(ctx context.Context, taskID uuid.UUID) ([]*models.TaskComment, error) {
	return d.Delegate.ListTaskComments(ctx, taskID)
}
func (d *TaskCommentStoreDecorator) UpdateTaskComment(ctx context.Context, comment *models.TaskComment) (*models.TaskComment, error) {
	return d.Delegate.UpdateTaskComment(ctx, comment)
}
func (d *TaskCommentStoreDecorator) DeleteTaskComment(ctx context.Context, id uuid.UUID) error {
	return d.Delegate.DeleteTaskComment(ctx, id)
}
func (d *TaskCommentStoreDecorator) WithTx(db database.Dbx) *DbTaskCommentStore {
	return d.Delegate.(*DbTaskCommentStore).WithTx(db)
}
