package tools

import "regexp"

var docIDPattern = regexp.MustCompile(`/d/([a-zA-Z0-9_-]+)`)

// ResolveID extracts a document/spreadsheet ID from a Google Workspace URL.
// If the input is not a URL, it is returned as-is.
func ResolveID(input string) string {
	matches := docIDPattern.FindStringSubmatch(input)
	if len(matches) >= 2 {
		return matches[1]
	}
	return input
}
