package tools

import (
	"regexp"
	"strings"
)

var docIDPattern = regexp.MustCompile(`^https?://docs\.google\.com/.+/d/([a-zA-Z0-9_-]+)`)

// ResolveID extracts a document/spreadsheet ID from an input string.
// If the input is a Google Workspace URL containing a /d/<id> path segment,
// the ID is extracted and returned. Otherwise, the input is returned unchanged.
func ResolveID(input string) string {
	input = strings.TrimSpace(input)
	matches := docIDPattern.FindStringSubmatch(input)
	if len(matches) >= 2 {
		return matches[1]
	}
	return input
}
