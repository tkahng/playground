package test

import "testing"

func SkipIfShort(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long running test in short mode")
	}
}
