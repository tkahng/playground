package routine_test

import (
	"sync"
	"testing"

	"github.com/tkahng/playground/internal/tools/routine"
)

func TestFireAndForget(t *testing.T) {
	called := false

	fn := func() {
		called = true
		panic("test")
	}

	wg := &sync.WaitGroup{}

	routine.FireAndForget(fn, wg)

	wg.Wait()

	if !called {
		t.Error("Expected fn to be called.")
	}
}
