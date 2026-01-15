package shared

import (
	"time"

	"github.com/google/uuid"
)

type AiUsage struct {
	ID               uuid.UUID `db:"id,pk" json:"id"`
	UserID           uuid.UUID `db:"user_id" json:"user_id"`
	PromptTokens     int64     `db:"prompt_tokens" json:"prompt_tokens"`
	CompletionTokens int64     `db:"completion_tokens" json:"completion_tokens"`
	TotalTokens      int64     `db:"total_tokens" json:"total_tokens"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time `db:"updated_at" json:"updated_at"`
}

type AiUsageListFilter struct {
	Q      string   `query:"q,omitempty" required:"false"`
	UserID string   `query:"user_id,omitempty" required:"false" format:"uuid"`
	Ids    []string `query:"ids,omitempty" required:"false" minimum:"1" maximum:"100" format:"uuid"`
}

type AiUsageListParams struct {
	PaginatedInput
	AiUsageListFilter
	SortParams
	Expand []string `query:"expand,omitempty" required:"false" minimum:"1" maximum:"100" enum:"subtasks"`
}
