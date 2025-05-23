package models

import (
	"time"

	"github.com/google/uuid"
)

// 'pending', 'processing', 'done', 'failed'
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusDone       JobStatus = "done"
	JobStatusFailed     JobStatus = "failed"
)

type Job struct {
	_           struct{}       `db:"jobs" json:"-"`
	ID          uuid.UUID      `db:"id" json:"id"`
	Type        string         `db:"type" json:"type"`
	UniqueKey   *string        `db:"unique_key" json:"unique_key"`
	Payload     map[string]any `db:"payload" json:"payload"`
	Status      JobStatus      `db:"status" json:"status"`
	RunAfter    time.Time      `db:"run_after" json:"run_after"`
	Attempts    int64          `db:"attempts" json:"attempts"`
	MaxAttempts int64          `db:"max_attempts" json:"max_attempts"`
	LastError   *string        `db:"last_error" json:"last_error"`
	CreatedAt   time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at" json:"updated_at"`
}
