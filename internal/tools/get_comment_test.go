package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	drive "google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
)

func TestGetComment_ReturnsCommentWithReplies(t *testing.T) {
	mock := &mockDriveClient{
		getCommentFn: func(_ context.Context, _, _ string, _ string, _ bool) (*drive.Comment, error) {
			return &drive.Comment{
				Id:                "c1",
				Author:            &drive.User{DisplayName: "Alice"},
				Content:           "Main comment",
				QuotedFileContent: &drive.CommentQuotedFileContent{Value: "quoted text"},
				Resolved:          false,
				CreatedTime:       "2026-02-23T00:00:00Z",
				Replies: []*drive.Reply{
					{
						Id:          "r1",
						Author:      &drive.User{DisplayName: "Bob"},
						Content:     "Reply one",
						CreatedTime: "2026-02-23T01:00:00Z",
					},
					{
						Id:          "r2",
						Author:      &drive.User{DisplayName: "Carol"},
						Content:     "Reply two",
						CreatedTime: "2026-02-23T02:00:00Z",
					},
				},
			}, nil
		},
	}

	result := getComment(context.Background(), mock, getCommentInput{
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
	mock := &mockDriveClient{
		getCommentFn: func(_ context.Context, _, _ string, _ string, _ bool) (*drive.Comment, error) {
			return &drive.Comment{
				Id:          "c1",
				Content:     "Comment without author",
				Resolved:    false,
				CreatedTime: "2026-02-23T00:00:00Z",
				Replies:     []*drive.Reply{},
			}, nil
		},
	}

	result := getComment(context.Background(), mock, getCommentInput{
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
	mock := &mockDriveClient{
		getCommentFn: func(_ context.Context, _, _ string, _ string, _ bool) (*drive.Comment, error) {
			return nil, &googleapi.Error{Code: 404, Message: "Comment not found."}
		},
	}

	result := getComment(context.Background(), mock, getCommentInput{
		DocumentID: "doc-id",
		CommentID:  "nonexistent-comment",
	})
	assert.True(t, result.IsError)
	assert.NotEmpty(t, result.Content)
}

func TestGetComment_URLInput_ExtractsID(t *testing.T) {
	const expectedID = "1A2B3C4D5E6F7G8H9I0J"

	mock := &mockDriveClient{
		getCommentFn: func(_ context.Context, fileID, _ string, _ string, _ bool) (*drive.Comment, error) {
			assert.Equal(t, expectedID, fileID)
			return &drive.Comment{
				Id:          "c1",
				Content:     "test",
				CreatedTime: "2026-01-01T00:00:00Z",
				Replies:     []*drive.Reply{},
			}, nil
		},
	}

	result := getComment(context.Background(), mock, getCommentInput{
		DocumentID: "https://docs.google.com/document/d/" + expectedID + "/edit",
		CommentID:  "c1",
	})
	assert.False(t, result.IsError)
}
