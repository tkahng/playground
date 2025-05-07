package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/tkahng/authgo/internal/tools/types"
)

type AuditLog struct {
	ID           uuid.UUID          `db:"id,pk" json:"id"`
	IP           *string            `db:"ip" json:"ip,omitempty"`
	Email        *string            `db:"email" json:"email,omitempty"`
	AuditLog     string             `db:"audit_log" json:"audit_log"`
	Attributes   types.JSONMap[any] `db:"attributes" json:"attributes"`
	CreationDate time.Time          `db:"creation_date" json:"creation_date"`
	Modification time.Time          `db:"modification_date" json:"modification_date"`
}
