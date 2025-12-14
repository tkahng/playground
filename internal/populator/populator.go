package populator

import (
	"context"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/database/repository"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/memo"
)

type Populator interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetMemberByID(ctx context.Context, id uuid.UUID) (*models.TeamMember, error)
	GetTeamByID(ctx context.Context, id uuid.UUID) (*models.Team, error)

	GetTaskByID(ctx context.Context, id uuid.UUID) (*models.Task, error)
	GetProjectByID(ctx context.Context, id uuid.UUID) (*models.TaskProject, error)
	GetParticipantByID(ctx context.Context, id uuid.UUID) (*models.RpsParticipant, error)
}

type DbPopulator struct {
	user        *memo.MemoizedStore[*models.User, uuid.UUID]
	member      *memo.MemoizedStore[*models.TeamMember, uuid.UUID]
	team        *memo.MemoizedStore[*models.Team, uuid.UUID]
	task        *memo.MemoizedStore[*models.Task, uuid.UUID]
	project     *memo.MemoizedStore[*models.TaskProject, uuid.UUID]
	participant *memo.MemoizedStore[*models.RpsParticipant, uuid.UUID]
}

func (s *DbPopulator) GetParticipantByID(ctx context.Context, id uuid.UUID) (*models.RpsParticipant, error) {
	return s.participant.Get(ctx, id)
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

func New(adapter stores.StorageAdapterInterface) Populator {
	return &DbPopulator{
		participant: memo.New(
			func(ctx context.Context, key uuid.UUID) (*models.RpsParticipant, error) {
				return adapter.Gaming().FindRpsParticipant(ctx, &stores.RpsParticipantFilter{
					Ids: []uuid.UUID{key},
				})
			},
			func(ctx context.Context, keys ...uuid.UUID) ([]*models.RpsParticipant, error) {
				return adapter.Gaming().FindRpsParticipants(ctx, &stores.RpsParticipantFilter{
					Ids: keys,
					PaginatedInput: repository.PaginatedInput{
						Page:    0,
						PerPage: 50,
					},
				})
			},
			func(rp *models.RpsParticipant) uuid.UUID {
				return rp.ID
			},
		),
		user: memo.New(
			func(ctx context.Context, key uuid.UUID) (*models.User, error) {
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

// func PopulateRpsGame(ctx context.Context, populator Populator, rpsGame *models.RpsGame) error {
// 	var requestedParticipant, invitedParticipant *models.RpsParticipant
// 	requestedParticipant, err := populator.GetParticipantByID(rpsGame.)
// 	return nil
// }

func PopulateTask(ctx context.Context, populator Populator, task *models.Task) error {
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
	team, err := populator.GetTeamByID(ctx, task.TeamID)
	if err != nil {
		return err
	}
	task.Team = team

	project, err := populator.GetProjectByID(ctx, task.ProjectID)
	if err != nil {
		return err
	}
	task.Project = project

	return nil
}

func PopulateTeamMember(ctx context.Context, populator Populator, member *models.TeamMember) error {
	if member.User == nil {
		if member.UserID != nil {
			user, err := populator.GetUserByID(ctx, *member.UserID)
			if err != nil {
				return err
			}
			member.User = user
		}
	}
	if member.Team == nil {
		team, err := populator.GetTeamByID(ctx, member.TeamID)
		if err != nil {
			return err
		}
		member.Team = team
	}
	return nil
}
