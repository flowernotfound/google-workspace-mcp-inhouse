package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	drive "google.golang.org/api/drive/v3"
)

const infoFields = "id,name,description,createdTime,modifiedTime,owners(displayName),lastModifyingUser(displayName),webViewLink"

// documentInfo is the JSON response shape for get_document_info.
type documentInfo struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	CreatedTime       string   `json:"created_time"`
	ModifiedTime      string   `json:"modified_time"`
	Owners            []string `json:"owners"`
	LastModifyingUser string   `json:"last_modifying_user"`
	WebViewLink       string   `json:"web_view_link"`
}

func getDocumentInfo(ctx context.Context, driveService *drive.Service, input getDocumentInfoInput) (*mcp.CallToolResult, error) {
	f, err := driveService.Files.Get(input.DocumentID).
		Fields(infoFields).
		Context(ctx).
		Do()
	if err != nil {
		return errorResult(err), nil
	}

	owners := make([]string, 0, len(f.Owners))
	for _, o := range f.Owners {
		owners = append(owners, o.DisplayName)
	}

	lastModifyingUser := ""
	if f.LastModifyingUser != nil {
		lastModifyingUser = f.LastModifyingUser.DisplayName
	}

	info := documentInfo{
		ID:                f.Id,
		Name:              f.Name,
		Description:       f.Description,
		CreatedTime:       f.CreatedTime,
		ModifiedTime:      f.ModifiedTime,
		Owners:            owners,
		LastModifyingUser: lastModifyingUser,
		WebViewLink:       f.WebViewLink,
	}

	data, err := json.Marshal(info)
	if err != nil {
		return errorResult(fmt.Errorf("failed to serialize response: %w", err)), nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, nil
}
