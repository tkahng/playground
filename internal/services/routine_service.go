package services

import (
	"sync"

	"github.com/tkahng/authgo/internal/tools/routine"
)

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

type RoutineServiceDecorator struct {
	Delegate          RoutineService
	Wg                *sync.WaitGroup
	FireAndForgetFunc func(f func(), wg ...*sync.WaitGroup)
}

func NewRoutineServiceDecorator() *RoutineServiceDecorator {
	return &RoutineServiceDecorator{
		Delegate: NewRoutineService(),
	}
}

func (r *RoutineServiceDecorator) FireAndForget(f func()) {
	routine.FireAndForget(f, r.Wg)
}
