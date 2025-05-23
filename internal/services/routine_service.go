package services

import "github.com/tkahng/authgo/internal/tools/routine"

type RoutineService interface {
	FireAndForget(f func())
}

type routineService struct {
}

func NewRoutineService() RoutineService {
	return &routineService{}
}

func (w *routineService) FireAndForget(f func()) {
	routine.FireAndForget(f)
}
