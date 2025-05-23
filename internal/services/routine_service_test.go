package services

import (
	"sync"

	"github.com/tkahng/authgo/internal/tools/routine"
)

type MockRoutineService struct{ Wg *sync.WaitGroup }

var _ RoutineService = (*MockRoutineService)(nil)

func (m *MockRoutineService) FireAndForget(f func()) {
	routine.FireAndForget(f, m.Wg)
}
