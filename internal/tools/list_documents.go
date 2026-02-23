package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	drive "google.golang.org/api/drive/v3"
)

const (
	defaultListMaxResults = 20
	maxListMaxResults     = 100
	defaultOrderBy        = "modifiedTime desc"
	docsFileMimeType      = "application/vnd.google-apps.document"
	listFields            = "nextPageToken,files(id,name,createdTime,modifiedTime,owners(displayName),webViewLink)"
)

// documentItem is the JSON response shape for a single document entry.
type documentItem struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	CreatedTime  string   `json:"created_time"`
	ModifiedTime string   `json:"modified_time"`
	Owners       []string `json:"owners"`
	WebViewLink  string   `json:"web_view_link"`
}

func listDocuments(ctx context.Context, driveService *drive.Service, input listDocumentsInput) (*mcp.CallToolResult, error) {
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

	q := fmt.Sprintf("mimeType='%s' and trashed=false", docsFileMimeType)
	if input.FolderID != nil && *input.FolderID != "" {
		q += fmt.Sprintf(" and '%s' in parents", escapeDriveQuery(*input.FolderID))
	}

	req := driveService.Files.List().
		Q(q).
		OrderBy(orderBy).
		PageSize(int64(maxResults)).
		Fields(listFields).
		Context(ctx)

	resp, err := req.Do()
	if err != nil {
		return errorResult(err), nil
	}

	items := make([]documentItem, 0, len(resp.Files))
	for _, f := range resp.Files {
		items = append(items, fileToDocumentItem(f))
	}

	data, err := json.Marshal(items)
	if err != nil {
		return errorResult(fmt.Errorf("failed to serialize response: %w", err)), nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, nil
}

// fileToDocumentItem converts a Drive file to a documentItem.
func fileToDocumentItem(f *drive.File) documentItem {
	owners := make([]string, 0, len(f.Owners))
	for _, o := range f.Owners {
		owners = append(owners, o.DisplayName)
	}
	return documentItem{
		ID:           f.Id,
		Name:         f.Name,
		CreatedTime:  f.CreatedTime,
		ModifiedTime: f.ModifiedTime,
		Owners:       owners,
		WebViewLink:  f.WebViewLink,
	}
}
