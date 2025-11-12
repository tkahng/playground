package stores

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/tools/memo"
)

type Populator interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetMemberByID(ctx context.Context, id uuid.UUID) (*models.TeamMember, error)
	GetTeamByID(ctx context.Context, id uuid.UUID) (*models.Team, error)

	GetTaskByID(ctx context.Context, id uuid.UUID) (*models.Task, error)
	GetProjectByID(ctx context.Context, id uuid.UUID) (*models.TaskProject, error)
}

type DbPopulator struct {
	user    *memo.MemoizedStore[*models.User, uuid.UUID]
	member  *memo.MemoizedStore[*models.TeamMember, uuid.UUID]
	team    *memo.MemoizedStore[*models.Team, uuid.UUID]
	task    *memo.MemoizedStore[*models.Task, uuid.UUID]
	project *memo.MemoizedStore[*models.TaskProject, uuid.UUID]
}

// GetMemberByID implements Populator.
func (s *DbPopulator) GetMemberByID(ctx context.Context, id uuid.UUID) (*models.TeamMember, error) {
	return s.member.Get(ctx, id)
}

// GetTeamByID implements Populator.
func (s *DbPopulator) GetTeamByID(ctx context.Context, id uuid.UUID) (*models.Team, error) {
	return s.team.Get(ctx, id)
}

// GetUserByID implements Populator.
func (s *DbPopulator) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return s.user.Get(ctx, id)
}

func (s *DbPopulator) GetTaskByID(ctx context.Context, id uuid.UUID) (*models.Task, error) {
	return s.task.Get(ctx, id)
}

func (s *DbPopulator) GetProjectByID(ctx context.Context, id uuid.UUID) (*models.TaskProject, error) {
	return s.project.Get(ctx, id)
}

func NewPopulator(adapter StorageAdapterInterface) Populator {
	return &DbPopulator{
		user: memo.NewMemoizedStore(func(ctx context.Context, key uuid.UUID) (*models.User, error) {
			return adapter.User().FindUserByID(ctx, key)
		}),
		member: memo.NewMemoizedStore(func(ctx context.Context, key uuid.UUID) (*models.TeamMember, error) {
			return adapter.TeamMember().FindTeamMember(ctx, &TeamMemberFilter{
				Ids: []uuid.UUID{key},
			})
		}),
		team: memo.NewMemoizedStore(func(ctx context.Context, key uuid.UUID) (*models.Team, error) {
			return adapter.TeamGroup().FindTeamByID(ctx, key)
		}),
		task: memo.NewMemoizedStore(func(ctx context.Context, key uuid.UUID) (*models.Task, error) {
			return adapter.Task().FindTaskByID(ctx, key)
		}),
		project: memo.NewMemoizedStore(func(ctx context.Context, key uuid.UUID) (*models.TaskProject, error) {
			return adapter.Task().FindTaskProjectByID(ctx, key)
		}),
	}

}

func getMember(ctx context.Context, populator Populator, memberId uuid.UUID) (*models.TeamMember, error) {
	member, err := populator.GetMemberByID(ctx, memberId)
	if err != nil {
		return nil, err
	}
	if member != nil {
		if member.UserID != nil {
			user, err := populator.GetUserByID(ctx, *member.UserID)
			if err != nil {
				return nil, err
			}
			member.User = user
		}
		return member, nil
	}
	return nil, nil
}
func PopulateTask(ctx context.Context, adapter StorageAdapterInterface, task *models.Task) error {
	populator := NewPopulator(adapter)
	if task.CreatedByMemberID != nil {
		member, err := getMember(ctx, populator, *task.CreatedByMemberID)
		if err != nil {
			return err
		}
		task.CreatedByMember = member
	}
	if task.AssigneeID != nil {
		member, err := getMember(ctx, populator, *task.AssigneeID)
		if err != nil {
			return err
		}
		task.Assignee = member
	}
	if task.ReporterID != nil {
		member, err := getMember(ctx, populator, *task.ReporterID)
		if err != nil {
			return err
		}
		task.Reporter = member
	}

	if task.ParentID != nil {
		parentTask, err := populator.GetTaskByID(ctx, *task.ParentID)
		if err != nil {
			return err
		}
		task.Parent = parentTask
	}

	return nil
}
