package urlshortner

import (
	"time"

	"github.com/google/uuid"
)

type ShortUrl struct {
	_         struct{}  `db:"short_urls" json:"-"`
	ID        uuid.UUID `db:"id" json:"id"`
	ShortCode string    `db:"short_code" json:"short_code"`
	SourceUrl string    `db:"source_url" json:"source_url"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
