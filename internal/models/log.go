package models

import (
	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/tools/types"
)

type Log struct {
	_         struct{}           `db:"logs" json:"-"`
	ID        uuid.UUID          `db:"id,pk" json:"id"`
	Message   string             `db:"message" json:"message"`
	Data      types.JSONMap[any] `db:"data" json:"data"`
	Timestamp string             `db:"timestamp" json:"timestamp"`
	CreatedAt string             `db:"created_at" json:"created_at"`
	UpdatedAt string             `db:"updated_at" json:"updated_at"`
}
