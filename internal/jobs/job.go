package jobs

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

type JobRow struct {
	_           struct{}  `db:"jobs" json:"-"`
	ID          uuid.UUID `db:"id" json:"id"`
	Kind        string    `db:"kind" json:"kind"`
	UniqueKey   *string   `db:"unique_key" json:"unique_key"`
	Payload     []byte    `db:"payload" json:"payload"`
	Status      JobStatus `db:"status" json:"status"`
	RunAfter    time.Time `db:"run_after" json:"run_after"`
	Attempts    int64     `db:"attempts" json:"attempts"`
	MaxAttempts int64     `db:"max_attempts" json:"max_attempts"`
	LastError   *string   `db:"last_error" json:"last_error"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// Job represents a single unit of work, holding both the arguments and
// information for a job with args of type T.
type Job[T JobArgs] struct {
	*JobRow

	// Args are the arguments for the job.
	Args T
}

type JobArgs interface {
	// Kind is a string that uniquely identifies the type of job. This must be
	// provided on your job arguments struct. Jobs are identified by a string
	// instead of being based on type names so that previously inserted jobs
	// can be worked across deploys even if job/worker types are renamed.
	//
	// Kinds should be formatted without spaces like `my_custom_job`,
	// `mycustomjob`, or `my-custom-job`. Many special characters like colons,
	// dots, hyphens, and underscores are allowed, but those like spaces and
	// commas, which would interfere with UI functionality, are invalid.
	//
	// After initially deploying a job, it's generally not safe to rename its
	// kind (unless the database is completely empty) because River won't know
	// which worker should work the old kind. Job kinds can be renamed safely
	// over multiple deploys using the JobArgsWithKindAliases interface.
	Kind() string
}
