package apis

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/apierrors"
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

type WorkflowCreateParams struct {
	TeamID string                   `path:"team-id" required:"true" format:"uuid"`
	Body   WorkflowCreateRequestDTO `required:"true"`
}

type WorkflowUpdateParams struct {
	TeamID     string                   `path:"team-id" required:"true" format:"uuid"`
	WorkflowID string                   `path:"workflow-id" required:"true" format:"uuid"`
	Body       stores.UpdateWorkflowDTO `required:"true"`
}

type WorkflowCreateRequestDTO struct {
	AppliesTo   string  `json:"applies_to" required:"true" enum:"project,task"`
	Name        string  `json:"name" required:"true" minLength:"1"`
	Description *string `json:"description,omitempty" required:"false"`
}

type WorkflowResponse struct {
	Body *Workflow
}

type WorkflowStatusCreateParams struct {
	TeamID     string                         `path:"team-id" required:"true" format:"uuid"`
	WorkflowID string                         `path:"workflow-id" required:"true" format:"uuid"`
	Body       stores.CreateWorkflowStatusDTO `required:"true"`
}

type WorkflowStatusUpdateParams struct {
	TeamID           string                         `path:"team-id" required:"true" format:"uuid"`
	WorkflowID       string                         `path:"workflow-id" required:"true" format:"uuid"`
	WorkflowStatusID string                         `path:"workflow-status-id" required:"true" format:"uuid"`
	Body             stores.UpdateWorkflowStatusDTO `required:"true"`
}

type WorkflowStatusDeleteParams struct {
	TeamID           string `path:"team-id" required:"true" format:"uuid"`
	WorkflowID       string `path:"workflow-id" required:"true" format:"uuid"`
	WorkflowStatusID string `path:"workflow-status-id" required:"true" format:"uuid"`
}

type WorkflowStatusResponse struct {
	Body *WorkflowStatus
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

func (api *Api) TeamWorkflowCreateBind(humaApi huma.API) {
	huma.Register(
		humaApi,
		huma.Operation{
			OperationID: "workflow-create",
			Method:      http.MethodPost,
			Path:        "/teams/{team-id}/workflows",
			Summary:     "Workflow create",
			Description: "Create a team workflow",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusBadRequest, http.StatusForbidden, http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireTeamInfo(),
				middleware.RequireTeamPermission(api.App(), shared.TeamPermissionWorkflowManage),
			),
		},
		func(ctx context.Context, input *WorkflowCreateParams) (*WorkflowResponse, error) {
			teamInfo := contextstore.GetContextTeamInfo(ctx)
			if teamInfo == nil {
				return nil, huma.Error401Unauthorized("Unauthorized")
			}
			workflow, err := api.App().Adapter().Task().CreateWorkflow(ctx, &stores.CreateWorkflowDTO{
				TeamID:            teamInfo.Team.ID,
				CreatedByMemberID: &teamInfo.Member.ID,
				AppliesTo:         input.Body.AppliesTo,
				Name:              input.Body.Name,
				Description:       input.Body.Description,
			})
			if err != nil {
				return nil, err
			}
			return &WorkflowResponse{Body: fromModelWorkflow(workflow)}, nil
		},
	)
}

func (api *Api) TeamWorkflowUpdateBind(humaApi huma.API) {
	huma.Register(
		humaApi,
		huma.Operation{
			OperationID: "workflow-update",
			Method:      http.MethodPut,
			Path:        "/teams/{team-id}/workflows/{workflow-id}",
			Summary:     "Workflow update",
			Description: "Update a team workflow",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusBadRequest, http.StatusForbidden, http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireTeamInfo(),
				middleware.RequireTeamPermission(api.App(), shared.TeamPermissionWorkflowManage),
			),
		},
		func(ctx context.Context, input *WorkflowUpdateParams) (*WorkflowResponse, error) {
			teamInfo := contextstore.GetContextTeamInfo(ctx)
			if teamInfo == nil {
				return nil, huma.Error401Unauthorized("Unauthorized")
			}
			workflowID, err := uuid.Parse(input.WorkflowID)
			if err != nil {
				return nil, huma.Error400BadRequest("Invalid workflow id", err)
			}
			workflow, err := api.App().Adapter().Task().FindWorkflowByID(ctx, workflowID)
			if err != nil {
				return nil, err
			}
			if workflow == nil || workflow.TeamID != teamInfo.Team.ID {
				return nil, apierrors.NotFound("workflow not found")
			}
			workflow, err = api.App().Adapter().Task().UpdateWorkflow(ctx, workflowID, &input.Body)
			if err != nil {
				return nil, err
			}
			return &WorkflowResponse{Body: fromModelWorkflow(workflow)}, nil
		},
	)
}

func (api *Api) TeamWorkflowStatusCreateBind(humaApi huma.API) {
	huma.Register(
		humaApi,
		huma.Operation{
			OperationID: "workflow-status-create",
			Method:      http.MethodPost,
			Path:        "/teams/{team-id}/workflows/{workflow-id}/statuses",
			Summary:     "Workflow status create",
			Description: "Create a workflow status",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusBadRequest, http.StatusForbidden, http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireTeamInfo(),
				middleware.RequireTeamPermission(api.App(), shared.TeamPermissionWorkflowManage),
			),
		},
		func(ctx context.Context, input *WorkflowStatusCreateParams) (*WorkflowStatusResponse, error) {
			teamInfo := contextstore.GetContextTeamInfo(ctx)
			if teamInfo == nil {
				return nil, huma.Error401Unauthorized("Unauthorized")
			}
			workflowID, err := uuid.Parse(input.WorkflowID)
			if err != nil {
				return nil, huma.Error400BadRequest("Invalid workflow id", err)
			}
			workflow, err := api.App().Adapter().Task().FindWorkflowByID(ctx, workflowID)
			if err != nil {
				return nil, err
			}
			if workflow == nil {
				return nil, apierrors.NotFound("workflow not found")
			}
			if workflow.TeamID != teamInfo.Team.ID {
				return nil, apierrors.NotFound("workflow not found")
			}
			status, err := api.App().Adapter().Task().CreateWorkflowStatus(ctx, workflowID, &input.Body)
			if err != nil {
				return nil, err
			}
			return &WorkflowStatusResponse{Body: fromModelWorkflowStatus(status)}, nil
		},
	)
}

func (api *Api) TeamWorkflowStatusUpdateBind(humaApi huma.API) {
	huma.Register(
		humaApi,
		huma.Operation{
			OperationID: "workflow-status-update",
			Method:      http.MethodPut,
			Path:        "/teams/{team-id}/workflows/{workflow-id}/statuses/{workflow-status-id}",
			Summary:     "Workflow status update",
			Description: "Update a workflow status",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusBadRequest, http.StatusForbidden, http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireTeamInfo(),
				middleware.RequireTeamPermission(api.App(), shared.TeamPermissionWorkflowManage),
			),
		},
		func(ctx context.Context, input *WorkflowStatusUpdateParams) (*WorkflowStatusResponse, error) {
			teamInfo := contextstore.GetContextTeamInfo(ctx)
			if teamInfo == nil {
				return nil, huma.Error401Unauthorized("Unauthorized")
			}
			workflowID, err := uuid.Parse(input.WorkflowID)
			if err != nil {
				return nil, huma.Error400BadRequest("Invalid workflow id", err)
			}
			workflowStatusID, err := uuid.Parse(input.WorkflowStatusID)
			if err != nil {
				return nil, huma.Error400BadRequest("Invalid workflow status id", err)
			}
			workflow, err := api.App().Adapter().Task().FindWorkflowByID(ctx, workflowID)
			if err != nil {
				return nil, err
			}
			if workflow == nil || workflow.TeamID != teamInfo.Team.ID {
				return nil, apierrors.NotFound("workflow not found")
			}
			status, err := api.App().Adapter().Task().FindWorkflowStatusByID(ctx, workflowStatusID)
			if err != nil {
				return nil, err
			}
			if status == nil || status.WorkflowID != workflow.ID {
				return nil, apierrors.NotFound("workflow status not found")
			}
			status, err = api.App().Adapter().Task().UpdateWorkflowStatus(ctx, workflowStatusID, &input.Body)
			if err != nil {
				return nil, err
			}
			return &WorkflowStatusResponse{Body: fromModelWorkflowStatus(status)}, nil
		},
	)
}

func (api *Api) TeamWorkflowStatusDeleteBind(humaApi huma.API) {
	huma.Register(
		humaApi,
		huma.Operation{
			OperationID: "workflow-status-delete",
			Method:      http.MethodDelete,
			Path:        "/teams/{team-id}/workflows/{workflow-id}/statuses/{workflow-status-id}",
			Summary:     "Workflow status delete",
			Description: "Delete an unused workflow status",
			Tags:        []string{"Task"},
			Errors:      []int{http.StatusBadRequest, http.StatusConflict, http.StatusForbidden, http.StatusNotFound},
			Security: []map[string][]string{{
				shared.BearerAuthSecurityKey: {},
			}},
			Middlewares: humamiddleware.HumaChiMiddlewares(
				middleware.RequireTeamInfo(),
				middleware.RequireTeamPermission(api.App(), shared.TeamPermissionWorkflowManage),
			),
		},
		func(ctx context.Context, input *WorkflowStatusDeleteParams) (*struct{}, error) {
			teamInfo := contextstore.GetContextTeamInfo(ctx)
			if teamInfo == nil {
				return nil, huma.Error401Unauthorized("Unauthorized")
			}
			workflowID, err := uuid.Parse(input.WorkflowID)
			if err != nil {
				return nil, huma.Error400BadRequest("Invalid workflow id", err)
			}
			workflowStatusID, err := uuid.Parse(input.WorkflowStatusID)
			if err != nil {
				return nil, huma.Error400BadRequest("Invalid workflow status id", err)
			}
			workflow, err := api.App().Adapter().Task().FindWorkflowByID(ctx, workflowID)
			if err != nil {
				return nil, err
			}
			if workflow == nil || workflow.TeamID != teamInfo.Team.ID {
				return nil, apierrors.NotFound("workflow not found")
			}
			status, err := api.App().Adapter().Task().FindWorkflowStatusByID(ctx, workflowStatusID)
			if err != nil {
				return nil, err
			}
			if status == nil || status.WorkflowID != workflow.ID {
				return nil, apierrors.NotFound("workflow status not found")
			}
			if err := api.App().Adapter().Task().DeleteWorkflowStatus(ctx, workflowStatusID); err != nil {
				return nil, err
			}
			return nil, nil
		},
	)
}
