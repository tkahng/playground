package jobs

import (
	"context"
)

// with the client using the AddWorker function.
type Worker[T JobArgs] interface {

	// Work performs the job and returns an error if the job failed. The context
	// will be configured with a timeout according to the worker settings and may
	// be cancelled for other reasons.
	//
	// If no error is returned, the job is assumed to have succeeded and will be
	// marked completed.
	//
	// It is important for any worker to respect context cancellation to enable
	// the client to respond to shutdown requests; there is no way to cancel a
	// running job that does not respect context cancellation, other than
	// terminating the process.
	Work(ctx context.Context, job *Job[T]) error
}
