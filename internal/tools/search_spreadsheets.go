package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func searchSpreadsheets(ctx context.Context, driveClient DriveClient, input searchSpreadsheetsInput) *mcp.CallToolResult {
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
		}
	}

	q := fmt.Sprintf(
		"mimeType='%s' and fullText contains '%s' and trashed=false",
		sheetsFileMimeType,
		escapeDriveQuery(trimmedQuery),
	)

	resp, err := driveClient.ListFiles(ctx, q, "", int64(maxResults), listFields)
	if err != nil {
		return errorResult(err)
	}

	items := make([]documentItem, 0, len(resp.Files))
	for _, f := range resp.Files {
		items = append(items, fileToDocumentItem(f))
	}

	data, err := json.Marshal(items)
	if err != nil {
		return errorResult(fmt.Errorf("failed to serialize response: %w", err))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}
}
