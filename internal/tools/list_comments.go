package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	drive "google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
)

// commentItem is the JSON response shape for a single comment.
type commentItem struct {
	ID          string `json:"id"`
	Author      string `json:"author"`
	Content     string `json:"content"`
	QuotedText  string `json:"quoted_text"`
	Resolved    bool   `json:"resolved"`
	CreatedTime string `json:"created_time"`
	ReplyCount  int    `json:"reply_count"`
}

func listComments(ctx context.Context, driveService *drive.Service, input listCommentsInput) (*mcp.CallToolResult, struct{}, error) {
	resp, err := driveService.Comments.List(input.DocumentID).
		Fields("comments(id,author,content,quotedFileContent,resolved,createdTime,replies)").
		IncludeDeleted(false).
		PageSize(100).
		Context(ctx).
		Do()
	if err != nil {
		return errorResult(err), struct{}{}, nil
	}

	items := make([]commentItem, 0, len(resp.Comments))
	for _, c := range resp.Comments {
		if !input.IncludeResolved && c.Resolved {
			continue
		}

		quotedText := ""
		if c.QuotedFileContent != nil {
			quotedText = c.QuotedFileContent.Value
		}

		author := ""
		if c.Author != nil {
			author = c.Author.DisplayName
		}

		items = append(items, commentItem{
			ID:          c.Id,
			Author:      author,
			Content:     c.Content,
			QuotedText:  quotedText,
			Resolved:    c.Resolved,
			CreatedTime: c.CreatedTime,
			ReplyCount:  len(c.Replies),
		})
	}

	data, err := json.Marshal(items)
	if err != nil {
		return errorResult(fmt.Errorf("failed to serialize response: %w", err)), struct{}{}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(data)},
		},
	}, struct{}{}, nil
}

// errorResult converts an error into an IsError CallToolResult.
func errorResult(err error) *mcp.CallToolResult {
	msg := err.Error()

	var apiErr *googleapi.Error
	if errors.As(err, &apiErr) {
		switch apiErr.Code {
		case 404:
			msg = fmt.Sprintf("document not found (HTTP 404): %s", apiErr.Message)
		case 403:
			msg = fmt.Sprintf("access denied (HTTP 403): %s", apiErr.Message)
		case 401:
			msg = "authentication failed (HTTP 401): run `google-workspace-mcp-inhouse auth` to re-authenticate"
		case 429:
			msg = "Google API rate limit exceeded (HTTP 429): please wait and try again"
		default:
			msg = fmt.Sprintf("Google API error (HTTP %d): %s", apiErr.Code, apiErr.Message)
		}
	}

	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			&mcp.TextContent{Text: msg},
		},
	}
}
