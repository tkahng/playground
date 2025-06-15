package shared

import (
	"time"

	"github.com/google/uuid"
)

type Media struct {
	ID        uuid.UUID `json:"id" db:"id" format:"uuid"`
	Filename  string    `json:"filename" db:"filename"`
	URL       string    `json:"url" db:"url" format:"uri"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
