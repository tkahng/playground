package apis

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/playground/internal/models"
	"github.com/tkahng/playground/internal/stores"
	"github.com/tkahng/playground/internal/tools/mapper"
)

type AdminAiUsageRecord struct {
	ID               uuid.UUID  `json:"id"`
	UserID           uuid.UUID  `json:"user_id"`
	TeamID           *uuid.UUID `json:"team_id"`
	TeamMemberID     *uuid.UUID `json:"team_member_id"`
	PromptTokens     int64      `json:"prompt_tokens"`
	CompletionTokens int64      `json:"completion_tokens"`
	TotalTokens      int64      `json:"total_tokens"`
	CreatedAt        time.Time  `json:"created_at"`
}

type AdminAiUsageListInput struct {
	stores.PaginatedInput
	TeamID *uuid.UUID `query:"team_id,omitempty" required:"false" format:"uuid"`
	Since  *time.Time `query:"since,omitempty" required:"false" format:"date-time"`
	Until  *time.Time `query:"until,omitempty" required:"false" format:"date-time"`
}

func (api *Api) AdminAiUsageList(ctx context.Context, input *AdminAiUsageListInput) (*struct {
	Body ApiPaginatedResponse[*AdminAiUsageRecord]
}, error) {
	filter := &stores.AiUsageFilter{
		PaginatedInput: input.PaginatedInput,
		TeamID:         input.TeamID,
		Since:          input.Since,
		Until:          input.Until,
	}

	rows, err := api.App().Adapter().AiUsage().ListAiUsages(ctx, filter)
	if err != nil {
		return nil, err
	}
	total, err := api.App().Adapter().AiUsage().CountAiUsages(ctx, filter)
	if err != nil {
		return nil, err
	}

	records := mapper.Map(rows, func(r *models.AiUsage) *AdminAiUsageRecord {
		return &AdminAiUsageRecord{
			ID:               r.ID,
			UserID:           r.UserID,
			TeamID:           r.TeamID,
			TeamMemberID:     r.TeamMemberID,
			PromptTokens:     r.PromptTokens,
			CompletionTokens: r.CompletionTokens,
			TotalTokens:      r.TotalTokens,
			CreatedAt:        r.CreatedAt,
		}
	})

	paginatedInput := &PaginatedInput{
		Page:    input.PaginatedInput.Page,
		PerPage: input.PaginatedInput.PerPage,
	}

	return &struct {
		Body ApiPaginatedResponse[*AdminAiUsageRecord]
	}{
		Body: ApiPaginatedResponse[*AdminAiUsageRecord]{
			Data: records,
			Meta: ApiGenerateMeta(paginatedInput, total),
		},
	}, nil
}
