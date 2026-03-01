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
		{
			name:  "non-Google URL containing /d/ pattern",
			input: "https://example.com/d/not-a-google-id/page",
			want:  "https://example.com/d/not-a-google-id/page",
		},
		{
			name:  "plain ID with leading and trailing spaces",
			input: "  abc123  ",
			want:  "abc123",
		},
		{
			name:  "URL with leading and trailing spaces",
			input: "  https://docs.google.com/document/d/abc123/edit  ",
			want:  "abc123",
		},
		{
			name:  "URL with empty ID after /d/",
			input: "https://docs.google.com/document/d//edit",
			want:  "https://docs.google.com/document/d//edit",
		},
		{
			name:  "plain text containing /d/ pattern",
			input: "some/d/fake-id/text",
			want:  "some/d/fake-id/text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveID(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}
