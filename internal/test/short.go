package test

import "testing"

func Short(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
}
