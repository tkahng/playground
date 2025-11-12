package stores

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/tools/memo"
)

type CallRecorder struct {
	called int
	mu     sync.RWMutex
}

func (r *CallRecorder) Called() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.called
}

func (r *CallRecorder) Increment() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.called++
}

type TestPopulator struct {
	Recorder *CallRecorder
	DbPopulator
}

func NewTestPopulator(adapter StorageAdapterInterface) Populator {
	recorder := &CallRecorder{
		mu:     sync.RWMutex{},
		called: 0,
	}

	return &TestPopulator{
		Recorder: recorder,
		DbPopulator: DbPopulator{
			user: memo.NewMemoizedStore(func(ctx context.Context, key uuid.UUID) (*models.User, error) {
				recorder.Increment()
				return adapter.User().FindUserByID(ctx, key)
			}),
			member: memo.NewMemoizedStore(func(ctx context.Context, key uuid.UUID) (*models.TeamMember, error) {
				recorder.Increment()
				return adapter.TeamMember().FindTeamMember(ctx, &TeamMemberFilter{
					Ids: []uuid.UUID{key},
				})
			}),
			team: memo.NewMemoizedStore(func(ctx context.Context, key uuid.UUID) (*models.Team, error) {
				recorder.Increment()
				return adapter.TeamGroup().FindTeamByID(ctx, key)
			}),
			task: memo.NewMemoizedStore(func(ctx context.Context, key uuid.UUID) (*models.Task, error) {
				recorder.Increment()
				return adapter.Task().FindTaskByID(ctx, key)
			}),
			project: memo.NewMemoizedStore(func(ctx context.Context, key uuid.UUID) (*models.TaskProject, error) {
				recorder.Increment()
				return adapter.Task().FindTaskProjectByID(ctx, key)
			}),
		},
	}

}
