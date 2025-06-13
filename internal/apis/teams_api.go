package apis

import (
	"context"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/contextstore"
	"github.com/tkahng/authgo/internal/models"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/stores"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

type CreateTeamInput struct {
	Name string `json:"name" required:"true"`
	Slug string `json:"slug" required:"true"`
}

type TeamOutput struct {
	Body *shared.Team `json:"body"`
}
type TeamInfoOutput struct {
	Body *shared.TeamInfo `json:"body"`
}

func (api *Api) CreateTeam(
	ctx context.Context,
	input *struct {
		Body CreateTeamInput `json:"body" required:"true"`
	},
) (
	*TeamOutput,
	error,
) {
	info := contextstore.GetContextUserInfo(ctx)
	if info == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}
	team, err := api.app.Team().CreateTeamWithOwner(
		ctx,
		input.Body.Name,
		input.Body.Slug,
		info.User.ID,
	)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, huma.Error500InternalServerError("team not found")
	}
	return &TeamOutput{
		Body: shared.FromTeamModel(&team.Team),
	}, nil
}

func (api *Api) CheckTeamSlug(
	ctx context.Context,
	input *struct {
		Body struct {
			Slug string `json:"slug" required:"true"`
		} `json:"body" required:"true"`
	},
) (
	*struct {
		Body struct {
			Exists bool `json:"exists"`
		}
	},
	error,
) {
	exists, err := api.app.Adapter().TeamGroup().CheckTeamSlug(ctx, input.Body.Slug)
	if err != nil {
		return nil, err
	}
	return &struct {
		Body struct {
			Exists bool "json:\"exists\""
		}
	}{
		Body: struct {
			Exists bool `json:"exists"`
		}{
			Exists: exists,
		},
	}, nil
}

func (api *Api) GetUserTeamMembers(
	ctx context.Context,
	input *shared.TeamMemberListInput,
) (
	*ApiPaginatedOutput[*shared.TeamMember],
	error,
) {
	info := contextstore.GetContextUserInfo(ctx)
	if info == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}
	teams, err := api.app.Team().FindTeamMembersByUserID(ctx, info.User.ID, input)
	if err != nil {
		return nil, err
	}
	if len(teams) == 0 {
		return nil, huma.Error500InternalServerError("teams not found")
	}
	count, err := api.app.Adapter().TeamMember().CountTeamMembers(ctx, &stores.TeamMemberFilter{
		UserIds: []uuid.UUID{info.User.ID},
	})
	if err != nil {
		return nil, err
	}
	return &ApiPaginatedOutput[*shared.TeamMember]{
		Body: ApiPaginatedResponse[*shared.TeamMember]{
			Data: mapper.Map(teams, shared.FromTeamMemberModel),
			Meta: GenerateMeta(&input.PaginatedInput, count),
		},
	}, nil
}

func (api *Api) GetUserTeams(
	ctx context.Context,
	input *shared.UserListTeamsParams,
) (
	*ApiPaginatedOutput[*shared.Team],
	error,
) {
	info := contextstore.GetContextUserInfo(ctx)
	if info == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}
	params := &stores.TeamFilter{
		UserIds: []uuid.UUID{info.User.ID},
	}
	if input != nil {
		params.Page = input.Page
		params.PerPage = input.PerPage
		params.SortBy = input.SortBy
		params.SortOrder = input.SortOrder
	}

	teams, err := api.app.Adapter().TeamGroup().ListTeams(ctx, params)
	if err != nil {
		return nil, err
	}
	if len(teams) > 0 {
		teamIds := mapper.Map(teams, func(t *models.Team) uuid.UUID {
			return t.ID
		})
		members, err := api.app.Adapter().TeamMember().LoadTeamMembersByUserAndTeamIds(ctx, info.User.ID, teamIds...)
		if err != nil {
			return nil, err
		}
		for idx := range teamIds {
			team := teams[idx]
			member := members[idx]
			if team != nil && member != nil {
				team.Members = append(team.Members, member)
			}
		}
	}
	count, err := api.app.Adapter().TeamGroup().CountTeams(ctx, params)
	if err != nil {
		return nil, err
	}
	return &ApiPaginatedOutput[*shared.Team]{
		Body: ApiPaginatedResponse[*shared.Team]{
			Data: mapper.Map(teams, shared.FromTeamModel),
			Meta: GenerateMeta(&input.PaginatedInput, count),
		},
	}, nil
}

func (api *Api) FindTeamInfoBySlug(
	ctx context.Context,
	input *struct {
		Slug string `path:"team-slug" required:"true"`
	},
) (
	*TeamInfoOutput,
	error,
) {
	info := contextstore.GetContextTeamInfo(ctx)
	if info == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}
	return &TeamInfoOutput{
		Body: &shared.TeamInfo{
			Team:   *shared.FromTeamModel(&info.Team),
			Member: *shared.FromTeamMemberModel(&info.Member),
			User:   *shared.FromUserModel(&info.User),
		},
	}, nil
}

func (api *Api) FindTeamMemberBySlug(
	ctx context.Context,
	input *struct {
		Slug string `path:"team-slug" required:"true"`
	},
) (
	*TeamMemberOutput,
	error,
) {
	info := contextstore.GetContextTeamInfo(ctx)
	if info == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}
	return &TeamMemberOutput{
		Body: shared.FromTeamMemberModel(&info.Member),
	}, nil
}

type TeamMemberOutput struct {
	Body *shared.TeamMember `json:"body"`
}

func (api *Api) GetActiveTeamMember(
	ctx context.Context,
	input *struct{},
) (
	*TeamMemberOutput,
	error,
) {
	info := contextstore.GetContextUserInfo(ctx)
	if info == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}
	team, err := api.app.Team().GetActiveTeamMember(ctx, info.User.ID)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, huma.Error404NotFound("team not found")
	}
	return &TeamMemberOutput{
		Body: shared.FromTeamMemberModel(team),
	}, nil
}

type UpdateTeamInput struct {
	TeamID string `path:"team-id" required:"true"`
	Body   struct {
		Name string `json:"name" required:"true"`
	} `json:"body" required:"true"`
}

func (api *Api) UpdateTeam(
	ctx context.Context,
	input *UpdateTeamInput,
) (
	*TeamOutput,
	error,
) {
	info := contextstore.GetContextTeamInfo(ctx)
	if info == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}
	team, err := api.app.Team().UpdateTeam(ctx, info.Team.ID, input.Body.Name)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, huma.Error500InternalServerError("team not found")
	}
	return &TeamOutput{
		Body: shared.FromTeamModel(team),
	}, nil
}

func (api *Api) DeleteTeam(
	ctx context.Context,
	input *struct {
		TeamID string `path:"team-id" required:"true"`
	},
) (
	*TeamOutput,
	error,
) {
	info := contextstore.GetContextTeamInfo(ctx)
	if info == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}
	slog.InfoContext(ctx, "Deleting team", slog.String("team_id", info.Team.ID.String()), slog.String("user_id", info.User.ID.String()))
	err := api.app.Team().DeleteTeam(ctx, info.Team.ID, info.User.ID)
	if err != nil {
		slog.ErrorContext(ctx, "error deleting team", "teamId", info.Team.ID.String(), "userId", info.User.ID.String(), "error", err)
		return nil, err
	}
	slog.InfoContext(ctx, "Team deleted successfully", slog.String("team_id", info.Team.ID.String()), slog.String("user_id", info.User.ID.String()))
	return nil, nil
}

func (api *Api) GetTeam(
	ctx context.Context,
	input *struct {
		TeamID string `path:"team-id" required:"true"`
	},
) (
	*TeamOutput,
	error,
) {
	info := contextstore.GetContextUserInfo(ctx)
	if info == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}
	uid, err := uuid.Parse(input.TeamID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid team ID")
	}
	team, err := api.app.Adapter().TeamGroup().FindTeamByID(ctx, uid)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, huma.Error404NotFound("team not found")
	}
	return &TeamOutput{
		Body: shared.FromTeamModel(team),
	}, nil
}

type SetCurrentTeamInput struct {
	TeamID string `json:"team_id" required:"true"`
}

func (api *Api) SetCurrentTeam(
	ctx context.Context,
	input *struct {
		Body SetCurrentTeamInput `json:"body" required:"true"`
	},
) (
	*struct {
		Body struct {
			Success bool `json:"success"`
		}
	},
	error,
) {
	info := contextstore.GetContextUserInfo(ctx)
	if info == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}
	passedTeamID, err := uuid.Parse(input.Body.TeamID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid team ID")
	}
	teamMember, err := api.app.Team().SetActiveTeamMember(ctx, info.User.ID, passedTeamID)
	if err != nil {
		return nil, err
	}
	if teamMember == nil {
		return nil, huma.Error404NotFound("team not found")
	}
	return &struct {
		Body struct {
			Success bool `json:"success"`
		}
	}{
		Body: struct {
			Success bool `json:"success"`
		}{
			Success: true,
		},
	}, nil
}
