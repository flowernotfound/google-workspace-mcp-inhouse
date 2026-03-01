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

func TestListComments_ReturnsComments(t *testing.T) {
	mock := &mockDriveClient{
		listCommentsFn: func(_ context.Context, _, _ string, _ bool, _ int64, _ string) (*drive.CommentList, error) {
			return &drive.CommentList{
				Comments: []*drive.Comment{
					{
						Id:                "c1",
						Author:            &drive.User{DisplayName: "Alice"},
						Content:           "First comment",
						QuotedFileContent: &drive.CommentQuotedFileContent{Value: "quoted text"},
						Resolved:          false,
						CreatedTime:       "2026-02-23T00:00:00Z",
						Replies:           []*drive.Reply{},
					},
					{
						Id:          "c2",
						Author:      &drive.User{DisplayName: "Bob"},
						Content:     "Second comment",
						Resolved:    false,
						CreatedTime: "2026-02-23T01:00:00Z",
						Replies:     []*drive.Reply{},
					},
				},
			}, nil
		},
	}

	result := listComments(context.Background(), mock, listCommentsInput{
		DocumentID:      "doc-id",
		IncludeResolved: false,
	})
	assert.False(t, result.IsError)

	var items []commentItem
	text := result.Content[0].(*mcp.TextContent).Text
	require.NoError(t, json.Unmarshal([]byte(text), &items))

	assert.Len(t, items, 2)
	assert.Equal(t, "c1", items[0].ID)
	assert.Equal(t, "Alice", items[0].Author)
	assert.Equal(t, "First comment", items[0].Content)
	assert.Equal(t, "quoted text", items[0].QuotedText)
	assert.Equal(t, "2026-02-23T00:00:00Z", items[0].CreatedTime)
}

func TestListComments_FilterResolved(t *testing.T) {
	mock := &mockDriveClient{
		listCommentsFn: func(_ context.Context, _, _ string, _ bool, _ int64, _ string) (*drive.CommentList, error) {
			return &drive.CommentList{
				Comments: []*drive.Comment{
					{
						Id:          "c1",
						Author:      &drive.User{DisplayName: "Alice"},
						Content:     "Open comment",
						Resolved:    false,
						CreatedTime: "2026-02-23T00:00:00Z",
						Replies:     []*drive.Reply{},
					},
					{
						Id:          "c2",
						Author:      &drive.User{DisplayName: "Bob"},
						Content:     "Resolved comment",
						Resolved:    true,
						CreatedTime: "2026-02-23T01:00:00Z",
						Replies:     []*drive.Reply{},
					},
				},
			}, nil
		},
	}

	result := listComments(context.Background(), mock, listCommentsInput{
		DocumentID:      "doc-id",
		IncludeResolved: false,
	})
	assert.False(t, result.IsError)

	var items []commentItem
	text := result.Content[0].(*mcp.TextContent).Text
	require.NoError(t, json.Unmarshal([]byte(text), &items))

	assert.Len(t, items, 1)
	assert.Equal(t, "c1", items[0].ID)
	assert.False(t, items[0].Resolved)
}

func TestListComments_IncludeResolved(t *testing.T) {
	mock := &mockDriveClient{
		listCommentsFn: func(_ context.Context, _, _ string, _ bool, _ int64, _ string) (*drive.CommentList, error) {
			return &drive.CommentList{
				Comments: []*drive.Comment{
					{
						Id:          "c1",
						Author:      &drive.User{DisplayName: "Alice"},
						Content:     "Open comment",
						Resolved:    false,
						CreatedTime: "2026-02-23T00:00:00Z",
						Replies:     []*drive.Reply{},
					},
					{
						Id:          "c2",
						Author:      &drive.User{DisplayName: "Bob"},
						Content:     "Resolved comment",
						Resolved:    true,
						CreatedTime: "2026-02-23T01:00:00Z",
						Replies:     []*drive.Reply{},
					},
				},
			}, nil
		},
	}

	result := listComments(context.Background(), mock, listCommentsInput{
		DocumentID:      "doc-id",
		IncludeResolved: true,
	})
	assert.False(t, result.IsError)

	var items []commentItem
	text := result.Content[0].(*mcp.TextContent).Text
	require.NoError(t, json.Unmarshal([]byte(text), &items))

	assert.Len(t, items, 2)
}

func TestListComments_Pagination(t *testing.T) {
	callCount := 0
	mock := &mockDriveClient{
		listCommentsFn: func(_ context.Context, _, _ string, _ bool, _ int64, pageToken string) (*drive.CommentList, error) {
			callCount++
			switch callCount {
			case 1:
				assert.Equal(t, "", pageToken)
				return &drive.CommentList{
					NextPageToken: "token-page2",
					Comments: []*drive.Comment{
						{
							Id:          "c1",
							Author:      &drive.User{DisplayName: "Alice"},
							Content:     "Comment one",
							Resolved:    false,
							CreatedTime: "2026-02-23T00:00:00Z",
							Replies:     []*drive.Reply{},
						},
						{
							Id:          "c2",
							Author:      &drive.User{DisplayName: "Bob"},
							Content:     "Comment two",
							Resolved:    false,
							CreatedTime: "2026-02-23T01:00:00Z",
							Replies:     []*drive.Reply{},
						},
					},
				}, nil
			case 2:
				assert.Equal(t, "token-page2", pageToken)
				return &drive.CommentList{
					Comments: []*drive.Comment{
						{
							Id:          "c3",
							Author:      &drive.User{DisplayName: "Carol"},
							Content:     "Comment three",
							Resolved:    false,
							CreatedTime: "2026-02-23T02:00:00Z",
							Replies:     []*drive.Reply{},
						},
					},
				}, nil
			default:
				t.Fatalf("unexpected call #%d to ListComments", callCount)
				return nil, nil
			}
		},
	}

	result := listComments(context.Background(), mock, listCommentsInput{
		DocumentID:      "doc-id",
		IncludeResolved: false,
	})
	assert.False(t, result.IsError)

	var items []commentItem
	text := result.Content[0].(*mcp.TextContent).Text
	require.NoError(t, json.Unmarshal([]byte(text), &items))

	assert.Equal(t, 2, callCount)
	assert.Len(t, items, 3)
	assert.Equal(t, "c1", items[0].ID)
	assert.Equal(t, "c2", items[1].ID)
	assert.Equal(t, "c3", items[2].ID)
}

func TestListComments_EmptyList(t *testing.T) {
	mock := &mockDriveClient{
		listCommentsFn: func(_ context.Context, _, _ string, _ bool, _ int64, _ string) (*drive.CommentList, error) {
			return &drive.CommentList{Comments: []*drive.Comment{}}, nil
		},
	}

	result := listComments(context.Background(), mock, listCommentsInput{
		DocumentID:      "doc-id",
		IncludeResolved: false,
	})
	assert.False(t, result.IsError)

	var items []commentItem
	text := result.Content[0].(*mcp.TextContent).Text
	require.NoError(t, json.Unmarshal([]byte(text), &items))
	assert.NotNil(t, items)
	assert.Empty(t, items)
}

func TestListComments_APIError(t *testing.T) {
	mock := &mockDriveClient{
		listCommentsFn: func(_ context.Context, _, _ string, _ bool, _ int64, _ string) (*drive.CommentList, error) {
			return nil, &googleapi.Error{Code: 404, Message: "File not found."}
		},
	}

	result := listComments(context.Background(), mock, listCommentsInput{
		DocumentID: "nonexistent-doc",
	})
	assert.True(t, result.IsError)
	assert.NotEmpty(t, result.Content)
}

func TestListComments_URLInput_ExtractsID(t *testing.T) {
	const expectedID = "1A2B3C4D5E6F7G8H9I0J"

	mock := &mockDriveClient{
		listCommentsFn: func(_ context.Context, fileID, _ string, _ bool, _ int64, _ string) (*drive.CommentList, error) {
			assert.Equal(t, expectedID, fileID)
			return &drive.CommentList{Comments: []*drive.Comment{}}, nil
		},
	}

	result := listComments(context.Background(), mock, listCommentsInput{
		DocumentID: "https://docs.google.com/document/d/" + expectedID + "/edit",
	})
	assert.False(t, result.IsError)
}
