package slug

import (
	"regexp"
	"strings"
)

// Remove any non-alphanumeric characters (except dashes)
var regex *regexp.Regexp = regexp.MustCompile("[^a-zA-Z0-9-]+")

// NewSlug creates a new Slug from a title string.
func NewSlug(title string) string {
	// Convert to lowercase
	slug := strings.ToLower(title)

	// Replace spaces with dashes
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = regex.ReplaceAllString(slug, "")

	return slug
}
