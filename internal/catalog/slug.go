package catalog

import (
	"strings"
	"unicode"
)

// Slugify builds a URL-safe slug, keeping Arabic letters intact so a project
// with no English title still gets a readable slug.
func Slugify(value string) string {
	var b strings.Builder
	lastDash := true // avoids a leading dash

	for _, r := range strings.ToLower(strings.TrimSpace(value)) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			lastDash = false
		case !lastDash:
			b.WriteRune('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}
