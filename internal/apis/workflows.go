package apis

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/contextstore"
	"github.com/tkahng/playground/internal/middleware"
	"github.com/tkahng/playground/internal/middleware/humamiddleware"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/shared"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/mapper"
	"github.com/tkahng/playground/internal/tools/utils"
)

type Workflow struct {
	ID                uuid.UUID         `json:"id"`
	TeamID            uuid.UUID         `json:"team_id"`
	CreatedByMemberID *uuid.UUID        `json:"created_by_member_id,omitempty" nullable:"true"`
	AppliesTo         string            `json:"applies_to"`
	Name              string            `json:"name"`
	Description       *string           `json:"description,omitempty" nullable:"true"`
	IsDefault         bool              `json:"is_default"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
	Statuses          []*WorkflowStatus `json:"statuses,omitempty"`
}

type WorkflowStatus struct {
	ID          uuid.UUID `json:"id"`
	WorkflowID  uuid.UUID `json:"workflow_id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description *string   `json:"description,omitempty" nullable:"true"`
	Category    string    `json:"category"`
	Color       *string   `json:"color,omitempty" nullable:"true"`
	Rank        float64   `json:"rank"`
	IsCompleted bool      `json:"is_completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func fromModelWorkflow(workflow *models.Workflow) *Workflow {
	if workflow == nil {
		return nil
	}
	return &Workflow{
		ID:                workflow.ID,
		TeamID:            workflow.TeamID,
		CreatedByMemberID: workflow.CreatedByMemberID,
		AppliesTo:         workflow.AppliesTo,
		Name:              workflow.Name,
		Description:       workflow.Description,
		IsDefault:         workflow.IsDefault,
		CreatedAt:         workflow.CreatedAt,
		UpdatedAt:         workflow.UpdatedAt,
		Statuses:          mapper.Map(workflow.Statuses, fromModelWorkflowStatus),
	}
}

func fromModelWorkflowStatus(status *models.WorkflowStatus) *WorkflowStatus {
	if status == nil {
		return nil
	}
	return &WorkflowStatus{
		ID:          status.ID,
		WorkflowID:  status.WorkflowID,
		Name:        status.Name,
		Slug:        status.Slug,
		Description: status.Description,
		Category:    status.Category,
		Color:       status.Color,
		Rank:        status.Rank,
		IsCompleted: status.IsCompleted,
		CreatedAt:   status.CreatedAt,
		UpdatedAt:   status.UpdatedAt,
	}
}

type WorkflowListParams struct {
	TeamID    string   `path:"team-id" required:"true" format:"uuid"`
	AppliesTo []string `query:"applies_to,omitempty" required:"false" minimum:"1" maximum:"10" enum:"project,task"`
	Ids       []string `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
}

type WorkflowListResponse struct {
	Body []*Workflow
}

func (api *Api) TeamWorkflowListBind(humaApi huma.API) {
	huma.Register(
		humaApi,
		huma.Operation{
			OperationID: "workflow-list",
			Method:      http.MethodGet,
			Path:        "/teams/{team-id}/workflows",
			Summary:     "Workflow list",
			Description: "List team workflows",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireTeamInfo(),
			),
		},
		func(ctx context.Context, input *WorkflowListParams) (*WorkflowListResponse, error) {
			teamInfo := contextstore.GetContextTeamInfo(ctx)
			if teamInfo == nil {
				return nil, huma.Error401Unauthorized("Unauthorized")
			}
			workflows, err := api.App().Adapter().Task().ListWorkflows(ctx, &stores.WorkflowFilter{
				TeamIds:   []uuid.UUID{teamInfo.Team.ID},
				Ids:       utils.ParseValidUUIDs(input.Ids...),
				AppliesTo: input.AppliesTo,
			})
			if err != nil {
				return nil, err
			}
			workflowIds := mapper.Map(workflows, func(workflow *models.Workflow) uuid.UUID {
				return workflow.ID
			})
			if len(workflowIds) == 0 {
				return &WorkflowListResponse{Body: []*Workflow{}}, nil
			}
			statuses, err := api.App().Adapter().Task().LoadWorkflowStatuses(ctx, workflowIds...)
			if err != nil {
				return nil, err
			}
			for idx, workflow := range workflows {
				workflow.Statuses = statuses[idx]
			}
			return &WorkflowListResponse{
				Body: mapper.Map(workflows, fromModelWorkflow),
			}, nil
		},
	)
}
