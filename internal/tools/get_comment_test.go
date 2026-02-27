package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetComment_ReturnsCommentWithReplies(t *testing.T) {
	mockResp := map[string]any{
		"id":                "c1",
		"author":            map[string]any{"displayName": "Alice"},
		"content":           "Main comment",
		"quotedFileContent": map[string]any{"value": "quoted text"},
		"resolved":          false,
		"createdTime":       "2026-02-23T00:00:00Z",
		"replies": []map[string]any{
			{
				"id":          "r1",
				"author":      map[string]any{"displayName": "Bob"},
				"content":     "Reply one",
				"createdTime": "2026-02-23T01:00:00Z",
			},
			{
				"id":          "r2",
				"author":      map[string]any{"displayName": "Carol"},
				"content":     "Reply two",
				"createdTime": "2026-02-23T02:00:00Z",
			},
		},
	}

	svc := newMockDriveService(t, jsonResponse(200, mockResp))
	result := getComment(context.Background(), svc, getCommentInput{
		DocumentID: "doc-id",
		CommentID:  "c1",
	})
	assert.False(t, result.IsError)

	var detail commentDetail
	text := result.Content[0].(*mcp.TextContent).Text
	require.NoError(t, json.Unmarshal([]byte(text), &detail))

	assert.Equal(t, "c1", detail.ID)
	assert.Equal(t, "Alice", detail.Author)
	assert.Equal(t, "Main comment", detail.Content)
	assert.Equal(t, "quoted text", detail.QuotedText)
	assert.False(t, detail.Resolved)
	assert.Equal(t, "2026-02-23T00:00:00Z", detail.CreatedTime)

	require.Len(t, detail.Replies, 2)
	assert.Equal(t, "r1", detail.Replies[0].ID)
	assert.Equal(t, "Bob", detail.Replies[0].Author)
	assert.Equal(t, "Reply one", detail.Replies[0].Content)
	assert.Equal(t, "2026-02-23T01:00:00Z", detail.Replies[0].CreatedTime)
	assert.Equal(t, "r2", detail.Replies[1].ID)
}

func TestGetComment_NilAuthorAndQuotedText(t *testing.T) {
	mockResp := map[string]any{
		"id":          "c1",
		"content":     "Comment without author",
		"resolved":    false,
		"createdTime": "2026-02-23T00:00:00Z",
		"replies":     []map[string]any{},
	}
	svc := newMockDriveService(t, jsonResponse(200, mockResp))
	result := getComment(context.Background(), svc, getCommentInput{
		DocumentID: "doc-id",
		CommentID:  "c1",
	})
	assert.False(t, result.IsError)

	var detail commentDetail
	text := result.Content[0].(*mcp.TextContent).Text
	require.NoError(t, json.Unmarshal([]byte(text), &detail))
	assert.Equal(t, "", detail.Author)
	assert.Equal(t, "", detail.QuotedText)
}

func TestGetComment_APIError(t *testing.T) {
	svc := newMockDriveService(t, googleAPIError(404, "Comment not found."))
	result := getComment(context.Background(), svc, getCommentInput{
		DocumentID: "doc-id",
		CommentID:  "nonexistent-comment",
	})
	assert.True(t, result.IsError)
	assert.NotEmpty(t, result.Content)
}
