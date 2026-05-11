package slug

import (
	"regexp"
	"strings"
)

var (
	nonAlphanumDash = regexp.MustCompile(`[^a-zA-Z0-9-]+`)
	consecutiveDash = regexp.MustCompile(`-{2,}`)
)

const MaxSlugLen = 50

// NewSlug creates a URL-safe slug from title.
// Consecutive dashes are collapsed, leading/trailing dashes are stripped,
// and the result is capped at MaxSlugLen characters.
func NewSlug(title string) string {
	s := strings.ToLower(title)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	s = nonAlphanumDash.ReplaceAllString(s, "")
	s = consecutiveDash.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > MaxSlugLen {
		s = s[:MaxSlugLen]
		s = strings.TrimRight(s, "-")
	}
	return s
}
