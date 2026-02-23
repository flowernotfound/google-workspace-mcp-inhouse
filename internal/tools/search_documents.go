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

// escapeDriveQuery escapes single quotes in a Drive API query string value.
func escapeDriveQuery(s string) string {
	return strings.ReplaceAll(s, `'`, `\'`)
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

	q := fmt.Sprintf(
		"mimeType='%s' and fullText contains '%s' and trashed=false",
		docsFileMimeType,
		escapeDriveQuery(input.Query),
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
		owners := make([]string, 0, len(f.Owners))
		for _, o := range f.Owners {
			owners = append(owners, o.DisplayName)
		}
		items = append(items, documentItem{
			ID:           f.Id,
			Name:         f.Name,
			CreatedTime:  f.CreatedTime,
			ModifiedTime: f.ModifiedTime,
			Owners:       owners,
			WebViewLink:  f.WebViewLink,
		})
	}

	data, err := json.Marshal(items)
	if err != nil {
		return errorResult(fmt.Errorf("failed to serialize response: %w", err)), struct{}{}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, struct{}{}, nil
}
