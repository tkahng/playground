package slug_test

import (
	"testing"

	"github.com/tkahng/playground/internal/tools/slug"
)

func TestNewSlug(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		title string
		want  string
	}{
		{
			name:  "Test Case 1",
			title: "test title",
			want:  "test-title",
		},
		{
			name:  "Test Case 2",
			title: "test-title-2",
			want:  "test-title-2",
		},
		{
			name:  "Test Case 3",
			title: "test_title_3",
			want:  "test-title-3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slug.NewSlug(tt.title)
			// TODO: update the condition below to compare got with tt.want.
			if got != tt.want {
				t.Errorf("NewSlug() = %v, want %v", got, tt.want)
			}
		})
	}
}
