package services

import (
	"sync"

	"github.com/tkahng/authgo/internal/tools/routine"
)

type mockRoutineService struct{ wg *sync.WaitGroup }

var _ WorkerService = (*mockRoutineService)(nil)

func (m *mockRoutineService) FireAndForget(f func()) {
	routine.FireAndForget(f, m.wg)
}
