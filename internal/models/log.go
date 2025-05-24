package models

import (
	"log/slog"

	"github.com/google/uuid"
)

type Log struct {
	_         struct{}   `db:"logs" json:"-"`
	ID        uuid.UUID  `db:"id,pk" json:"id"`
	Level     slog.Level `db:"level" json:"level"`
	Source    *string    `db:"source" json:"source"`
	Message   string     `db:"message" json:"message"`
	Data      []byte     `db:"data" json:"data"`
	CreatedAt string     `db:"created_at" json:"created_at"`
}

// id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
// level int NOT NULL DEFAULT 0,
// source text,
// message text NOT NULL,
// data jsonb NOT NULL,
// created_at timestamptz NOT NULL DEFAULT now()
