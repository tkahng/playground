package shared

import (
	"time"

	"github.com/google/uuid"
)

type MediaListFilter struct {
	Q      string `query:"q,omitempty" required:"false"`
	UserID string `query:"userId,omitempty" format:"uuid" required:"false"`
}

type MediaListParams struct {
	PaginatedInput
	MediaListFilter
	SortParams
}

type Media struct {
	ID        uuid.UUID `json:"id" db:"id" format:"uuid"`
	Filename  string    `json:"filename" db:"filename"`
	URL       string    `json:"url" db:"url" format:"uri"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
