package slug_test

import (
	"strings"
	"testing"

	"github.com/tkahng/playground/internal/tools/slug"
)

func TestNewSlug(t *testing.T) {
	tests := []struct {
		name  string
		title string
		want  string
	}{
		{name: "spaces become dashes", title: "test title", want: "test-title"},
		{name: "existing dashes preserved", title: "test-title-2", want: "test-title-2"},
		{name: "underscores become dashes", title: "test_title_3", want: "test-title-3"},
		{name: "leading/trailing spaces trimmed", title: " team ", want: "team"},
		{name: "consecutive spaces collapse to one dash", title: "team  name", want: "team-name"},
		{name: "special chars removed", title: "team!@#name", want: "teamname"},
		{name: "all special chars returns empty", title: "!!!###", want: ""},
		{name: "only dashes returns empty", title: "---", want: ""},
		{name: "mixed leading special chars trimmed", title: "!my team!", want: "my-team"},
		{
			name:  "long title truncated at MaxSlugLen",
			title: strings.Repeat("a", slug.MaxSlugLen+10),
			want:  strings.Repeat("a", slug.MaxSlugLen),
		},
		{
			name:  "truncation does not leave trailing dash",
			title: strings.Repeat("a", slug.MaxSlugLen-1) + " extra",
			want:  strings.Repeat("a", slug.MaxSlugLen-1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slug.NewSlug(tt.title)
			if got != tt.want {
				t.Errorf("NewSlug(%q) = %q, want %q", tt.title, got, tt.want)
			}
		})
	}
}
