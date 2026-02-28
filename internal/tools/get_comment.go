package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// replyItem is the JSON response shape for a single reply.
type replyItem struct {
	ID          string `json:"id"`
	Author      string `json:"author"`
	Content     string `json:"content"`
	CreatedTime string `json:"created_time"`
}

// commentDetail is the JSON response shape for get_comment.
type commentDetail struct {
	ID          string      `json:"id"`
	Author      string      `json:"author"`
	Content     string      `json:"content"`
	QuotedText  string      `json:"quoted_text"`
	Resolved    bool        `json:"resolved"`
	CreatedTime string      `json:"created_time"`
	Replies     []replyItem `json:"replies"`
}

func getComment(ctx context.Context, driveClient DriveClient, input getCommentInput) *mcp.CallToolResult {
	c, err := driveClient.GetComment(ctx, input.DocumentID, input.CommentID,
		"id,author,content,quotedFileContent,resolved,createdTime,replies(id,author,content,createdTime)", false)
	if err != nil {
		return errorResult(err)
	}

	quotedText := ""
	if c.QuotedFileContent != nil {
		quotedText = c.QuotedFileContent.Value
	}

	author := ""
	if c.Author != nil {
		author = c.Author.DisplayName
	}

	replies := make([]replyItem, 0, len(c.Replies))
	for _, r := range c.Replies {
		replyAuthor := ""
		if r.Author != nil {
			replyAuthor = r.Author.DisplayName
		}
		replies = append(replies, replyItem{
			ID:          r.Id,
			Author:      replyAuthor,
			Content:     r.Content,
			CreatedTime: r.CreatedTime,
		})
	}

	detail := commentDetail{
		ID:          c.Id,
		Author:      author,
		Content:     c.Content,
		QuotedText:  quotedText,
		Resolved:    c.Resolved,
		CreatedTime: c.CreatedTime,
		Replies:     replies,
	}

	data, err := json.Marshal(detail)
	if err != nil {
		return errorResult(fmt.Errorf("failed to serialize response: %w", err))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(data)},
		},
	}
}
