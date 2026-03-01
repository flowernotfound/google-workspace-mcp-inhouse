package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const sheetsFileMimeType = "application/vnd.google-apps.spreadsheet"

func listSpreadsheets(ctx context.Context, driveClient DriveClient, input listSpreadsheetsInput) *mcp.CallToolResult {
	maxResults := defaultListMaxResults
	if input.MaxResults != nil {
		maxResults = *input.MaxResults
		if maxResults > maxListMaxResults {
			maxResults = maxListMaxResults
		}
		if maxResults < 1 {
			maxResults = 1
		}
	}

	orderBy := defaultOrderBy
	if input.OrderBy != nil && *input.OrderBy != "" {
		orderBy = *input.OrderBy
	}

	q := fmt.Sprintf("mimeType='%s' and trashed=false", sheetsFileMimeType)
	if input.FolderID != nil && *input.FolderID != "" {
		q += fmt.Sprintf(" and '%s' in parents", escapeDriveQuery(*input.FolderID))
	}

	resp, err := driveClient.ListFiles(ctx, q, orderBy, int64(maxResults), listFields)
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
