package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Google Docs URL",
			input: "https://docs.google.com/document/d/1A2B3C4D5E6F7G8H9I0J/edit",
			want:  "1A2B3C4D5E6F7G8H9I0J",
		},
		{
			name:  "Google Sheets URL",
			input: "https://docs.google.com/spreadsheets/d/xyz789_ABC-def/edit#gid=0",
			want:  "xyz789_ABC-def",
		},
		{
			name:  "URL with query parameters",
			input: "https://docs.google.com/document/d/abc123/edit?usp=sharing",
			want:  "abc123",
		},
		{
			name:  "URL with fragment",
			input: "https://docs.google.com/spreadsheets/d/sheet-id-123/edit#gid=456",
			want:  "sheet-id-123",
		},
		{
			name:  "ID with special characters",
			input: "https://docs.google.com/document/d/1a-B_c2D/edit",
			want:  "1a-B_c2D",
		},
		{
			name:  "plain ID",
			input: "abc123",
			want:  "abc123",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "URL without /d/ pattern",
			input: "https://example.com/some/path",
			want:  "https://example.com/some/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveID(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
