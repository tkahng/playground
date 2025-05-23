package jobs

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"time"
)

func Execute[T JobArgs](ctx context.Context, worker Worker[T], job *Job[T]) (err error) {
	start := time.Now()
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			log.Printf("[worker:%s] panic: %v\n%s", job.Args.Kind(), r, stack)
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	log.Printf("[worker:%s] starting job %s", job.Args.Kind(), job.ID)
	err = worker.Work(ctx, job)
	if err != nil {
		log.Printf("[worker:%s] failed: %v", job.Args.Kind(), err)
	} else {
		log.Printf("[worker:%s] done in %s", job.Args.Kind(), time.Since(start))
	}
	return
}
