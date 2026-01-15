package models

import (
	"github.com/google/uuid"
)

type Log struct {
	_         struct{}  `db:"logs" schema:"app" json:"-"`
	ID        uuid.UUID `db:"id,pk" json:"id"`
	Level     int       `db:"level" json:"level"`
	Message   string    `db:"message" json:"message"`
	Data      []byte    `db:"data" json:"data"`
	CreatedAt string    `db:"created_at" json:"created_at"`
}
