package services

import "github.com/tkahng/authgo/internal/tools/routine"

type WorkerService interface {
	FireAndForget(f func())
}

type workerService struct {
}

func NewWorkerService() WorkerService {
	return &workerService{}
}

func (w *workerService) FireAndForget(f func()) {
	routine.FireAndForget(f)
}
