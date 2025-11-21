package populator

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/memo"
)

type callRecorder struct {
	called int
	mu     sync.RWMutex
}

func (r *callRecorder) Called() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.called
}

func (r *callRecorder) Increment() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.called++
}

type TestPopulator struct {
	Recorder *callRecorder
	DbPopulator
}

func NewTestPopulator(adapter stores.StorageAdapterInterface) *TestPopulator {
	recorder := &callRecorder{
		mu:     sync.RWMutex{},
		called: 0,
	}

	return &TestPopulator{
		Recorder: recorder,
		DbPopulator: DbPopulator{
			user: memo.New(
				func(ctx context.Context, key uuid.UUID) (*models.User, error) {
					recorder.Increment()
					return adapter.User().FindUserByID(ctx, key)
				},
				func(ctx context.Context, keys ...uuid.UUID) ([]*models.User, error) {
					return adapter.User().FindUsers(ctx, &stores.UserFilter{
						Ids: keys,
						PaginatedInput: stores.PaginatedInput{
							Page:    0,
							PerPage: 50,
						},
					})
				},
				func(u *models.User) uuid.UUID {
					return u.ID
				},
			),
			member: memo.New(
				func(ctx context.Context, key uuid.UUID) (*models.TeamMember, error) {
					recorder.Increment()
					return adapter.TeamMember().FindTeamMember(ctx, &stores.TeamMemberFilter{
						Ids: []uuid.UUID{key},
					})
				},
				func(ctx context.Context, keys ...uuid.UUID) ([]*models.TeamMember, error) {
					return adapter.TeamMember().FindTeamMembers(ctx, &stores.TeamMemberFilter{
						Ids: keys,
						PaginatedInput: stores.PaginatedInput{
							Page:    0,
							PerPage: 50,
						},
					})
				},
				func(m *models.TeamMember) uuid.UUID {
					return m.ID
				},
			),
			team: memo.New(
				func(ctx context.Context, key uuid.UUID) (*models.Team, error) {
					recorder.Increment()
					return adapter.TeamGroup().FindTeamByID(ctx, key)
				},
				func(ctx context.Context, keys ...uuid.UUID) ([]*models.Team, error) {
					return adapter.TeamGroup().ListTeams(ctx, &stores.TeamFilter{
						Ids: keys,
						PaginatedInput: stores.PaginatedInput{
							Page:    0,
							PerPage: 50,
						},
					})
				},
				func(t *models.Team) uuid.UUID {
					return t.ID
				},
			),
			task: memo.New(
				func(ctx context.Context, key uuid.UUID) (*models.Task, error) {
					recorder.Increment()
					return adapter.Task().FindTaskByID(ctx, key)
				},
				func(ctx context.Context, keys ...uuid.UUID) ([]*models.Task, error) {
					return adapter.Task().ListTasks(ctx, &stores.TaskFilter{
						Ids: keys,
						PaginatedInput: stores.PaginatedInput{
							Page:    0,
							PerPage: 50,
						},
					})
				},
				func(t *models.Task) uuid.UUID {
					return t.ID
				},
			),
			project: memo.New(
				func(ctx context.Context, key uuid.UUID) (*models.TaskProject, error) {
					recorder.Increment()
					return adapter.Task().FindTaskProjectByID(ctx, key)
				},
				func(ctx context.Context, keys ...uuid.UUID) ([]*models.TaskProject, error) {
					return adapter.Task().ListTaskProjects(ctx, &stores.TaskProjectsFilter{
						Ids: keys,
						PaginatedInput: stores.PaginatedInput{
							Page:    0,
							PerPage: 50,
						},
					})
				},
				func(p *models.TaskProject) uuid.UUID {
					return p.ID
				},
			),
		},
	}
}
