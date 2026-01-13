package test

import (
	"fmt"
	"testing"
	"time"
)

func TestRandomTimeBetween(t *testing.T) {
	t.Run("should return a time within the specified range", func(t *testing.T) {
		start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

		for i := 0; i < 100; i++ {
			got := RandomTimeBetween(start, end)
			// The generated time should be >= start and < end.
			if got.Before(start) || !got.Before(end) {
				t.Errorf("RandomTimeBetween() generated time %v, which is outside the expected range [%v, %v)", got, start, end)
			}
		}
	})

	t.Run("should panic if start time is not before end time", func(t *testing.T) {
		start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

		testCases := []struct {
			name      string
			startTime time.Time
			endTime   time.Time
		}{
			{
				name:      "start is after end",
				startTime: start.Add(time.Hour),
				endTime:   start,
			},
			{
				name:      "start is equal to end",
				startTime: start,
				endTime:   start,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				defer func() {
					r := recover()
					if r == nil {
						t.Errorf("The code did not panic as expected")
						return
					}
					expectedPanicMsg := "start time must be before end time"
					if fmt.Sprintf("%v", r) != expectedPanicMsg {
						t.Errorf("Expected panic message '%s', but got '%v'", expectedPanicMsg, r)
					}
				}()

				RandomTimeBetween(tc.startTime, tc.endTime)
			})
		}
	})
}
