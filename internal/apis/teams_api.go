package apis

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
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

func (api *Api) GetUserTeams(
	ctx context.Context,
	input *shared.PaginatedInput,
) (
	*shared.PaginatedOutput[*shared.TeamMember],
	error,
) {
	info := contextstore.GetContextUserInfo(ctx)
	if info == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}
	teams, err := api.app.Team().Store().FindTeamMembersByUserID(ctx, info.User.ID, input)
	if err != nil {
		return nil, err
	}
	if teams == nil {
		return nil, huma.Error500InternalServerError("teams not found")
	}
	count, err := api.app.Team().Store().CountTeamMembersByUserID(ctx, info.User.ID)
	if err != nil {
		return nil, err
	}
	return &shared.PaginatedOutput[*shared.TeamMember]{
		Body: shared.PaginatedResponse[*shared.TeamMember]{
			Data: mapper.Map(teams, shared.FromTeamMemberModel),
			Meta: shared.GenerateMeta(input, count),
		},
	}, nil

}
