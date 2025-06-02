package apis

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/contextstore"
	"github.com/tkahng/authgo/internal/shared"
	"github.com/tkahng/authgo/internal/tools/mapper"
)

type CreateTeamInput struct {
	Name string `json:"name" required:"true"`
	Slug string `json:"slug" required:"true"`
}

type TeamOutput struct {
	Body *shared.Team `json:"body"`
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
	team, err := api.app.Team().CreateTeam(
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
	exists, err := api.app.Team().Store().CheckTeamSlug(ctx, input.Body.Slug)
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
	*shared.PaginatedOutput[*shared.TeamMember],
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
	count, err := api.app.Team().Store().CountTeamMembersByUserID(ctx, info.User.ID)
	if err != nil {
		return nil, err
	}
	return &shared.PaginatedOutput[*shared.TeamMember]{
		Body: shared.PaginatedResponse[*shared.TeamMember]{
			Data: mapper.Map(teams, shared.FromTeamMemberModel),
			Meta: shared.GenerateMeta(&input.PaginatedInput, count),
		},
	}, nil
}

func (api *Api) GetUserTeams(
	ctx context.Context,
	input *shared.UserListTeamsParams,
) (
	*shared.PaginatedOutput[*shared.Team],
	error,
) {
	info := contextstore.GetContextUserInfo(ctx)
	if info == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}
	params := &shared.ListTeamsParams{
		ListTeamsFilter: shared.ListTeamsFilter{
			UserID: info.User.ID.String(),
		},
	}
	if input != nil {
		params.PaginatedInput = input.PaginatedInput
		params.SortParams = input.SortParams
	}

	teams, err := api.app.Team().Store().ListTeams(ctx, params)
	if err != nil {
		return nil, err
	}
	if len(teams) == 0 {
		return nil, huma.Error500InternalServerError("teams not found")
	}
	count, err := api.app.Team().Store().CountTeams(ctx, params)
	if err != nil {
		return nil, err
	}
	return &shared.PaginatedOutput[*shared.Team]{
		Body: shared.PaginatedResponse[*shared.Team]{
			Data: mapper.Map(teams, shared.FromTeamModel),
			Meta: shared.GenerateMeta(&input.PaginatedInput, count),
		},
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
	err := api.app.Team().DeleteTeam(ctx, info.Team.ID, *info.Member.UserID)
	if err != nil {
		return nil, err
	}
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
	team, err := api.app.Team().Store().FindTeamByID(ctx, uid)
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
