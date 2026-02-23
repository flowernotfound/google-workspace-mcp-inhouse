package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	drive "google.golang.org/api/drive/v3"
)

const (
	defaultSearchMaxResults = 10
	maxSearchMaxResults     = 50
)

// escapeDriveQuery escapes backslashes and single quotes in a Drive API query string value.
// Backslashes must be escaped before single quotes to avoid double-escaping.
func escapeDriveQuery(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\`), `'`, `\'`)
}

func searchDocuments(ctx context.Context, driveService *drive.Service, input searchDocumentsInput) (*mcp.CallToolResult, struct{}, error) {
	maxResults := defaultSearchMaxResults
	if input.MaxResults != nil {
		maxResults = *input.MaxResults
		if maxResults > maxSearchMaxResults {
			maxResults = maxSearchMaxResults
		}
		if maxResults < 1 {
			maxResults = 1
		}
	}

	trimmedQuery := strings.TrimSpace(input.Query)
	if trimmedQuery == "" {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "search query must not be empty"},
			},
		}, struct{}{}, nil
	}

	q := fmt.Sprintf(
		"mimeType='%s' and fullText contains '%s' and trashed=false",
		docsFileMimeType,
		escapeDriveQuery(trimmedQuery),
	)

	resp, err := driveService.Files.List().
		Q(q).
		PageSize(int64(maxResults)).
		Fields(listFields).
		Context(ctx).
		Do()
	if err != nil {
		return errorResult(err), struct{}{}, nil
	}

	items := make([]documentItem, 0, len(resp.Files))
	for _, f := range resp.Files {
		items = append(items, fileToDocumentItem(f))
	}

	data, err := json.Marshal(items)
	if err != nil {
		return errorResult(fmt.Errorf("failed to serialize response: %w", err)), struct{}{}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, struct{}{}, nil
}
