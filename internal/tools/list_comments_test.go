package tools

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	drive "google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// mockTransport is a http.RoundTripper that delegates to a function.
// Used to mock Google API responses without network access.
type mockTransport struct {
	fn func(*http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.fn(req)
}

// newMockDriveService creates a Drive service backed by a mock transport.
func newMockDriveService(t *testing.T, fn func(*http.Request) (*http.Response, error)) *drive.Service {
	t.Helper()
	svc, err := drive.NewService(context.Background(),
		option.WithHTTPClient(&http.Client{Transport: &mockTransport{fn: fn}}),
	)
	require.NoError(t, err)
	require.NoError(t, err)
	return svc
}

// jsonResponse returns a mock HTTP response with a JSON body.
func jsonResponse(statusCode int, body any) func(*http.Request) (*http.Response, error) {
	return func(_ *http.Request) (*http.Response, error) {
		data, err := json.Marshal(body)
		if err != nil {
			panic(err)
		}
		h := make(http.Header)
		h.Set("Content-Type", "application/json")
		return &http.Response{
			StatusCode: statusCode,
			Body:       io.NopCloser(strings.NewReader(string(data))),
			Header:     h,
		}, nil
	}
}

// googleAPIError returns a mock 4xx response in the Google API error envelope format.
func googleAPIError(code int, message string) func(*http.Request) (*http.Response, error) {
	body := map[string]any{
		"error": map[string]any{
			"code":    code,
			"message": message,
			"errors":  []map[string]any{{"message": message, "domain": "global", "reason": "notFound"}},
		},
	}
	return jsonResponse(code, body)
}

func TestListComments_ReturnsComments(t *testing.T) {
	mockResp := map[string]any{
		"kind": "drive#commentList",
		"comments": []map[string]any{
			{
				"id":                "c1",
				"author":            map[string]any{"displayName": "Alice"},
				"content":           "First comment",
				"quotedFileContent": map[string]any{"value": "quoted text"},
				"resolved":          false,
				"createdTime":       "2026-02-23T00:00:00Z",
				"replies":           []map[string]any{},
			},
			{
				"id":          "c2",
				"author":      map[string]any{"displayName": "Bob"},
				"content":     "Second comment",
				"resolved":    false,
				"createdTime": "2026-02-23T01:00:00Z",
				"replies":     []map[string]any{},
			},
		},
	}

	svc := newMockDriveService(t, jsonResponse(200, mockResp))
	result := listComments(context.Background(), svc, listCommentsInput{
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
	mockResp := map[string]any{
		"kind": "drive#commentList",
		"comments": []map[string]any{
			{
				"id":          "c1",
				"author":      map[string]any{"displayName": "Alice"},
				"content":     "Open comment",
				"resolved":    false,
				"createdTime": "2026-02-23T00:00:00Z",
				"replies":     []map[string]any{},
			},
			{
				"id":          "c2",
				"author":      map[string]any{"displayName": "Bob"},
				"content":     "Resolved comment",
				"resolved":    true,
				"createdTime": "2026-02-23T01:00:00Z",
				"replies":     []map[string]any{},
			},
		},
	}

	svc := newMockDriveService(t, jsonResponse(200, mockResp))
	result := listComments(context.Background(), svc, listCommentsInput{
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
	mockResp := map[string]any{
		"kind": "drive#commentList",
		"comments": []map[string]any{
			{
				"id":          "c1",
				"author":      map[string]any{"displayName": "Alice"},
				"content":     "Open comment",
				"resolved":    false,
				"createdTime": "2026-02-23T00:00:00Z",
				"replies":     []map[string]any{},
			},
			{
				"id":          "c2",
				"author":      map[string]any{"displayName": "Bob"},
				"content":     "Resolved comment",
				"resolved":    true,
				"createdTime": "2026-02-23T01:00:00Z",
				"replies":     []map[string]any{},
			},
		},
	}

	svc := newMockDriveService(t, jsonResponse(200, mockResp))
	result := listComments(context.Background(), svc, listCommentsInput{
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
	page1 := map[string]any{
		"nextPageToken": "token-page2",
		"comments": []map[string]any{
			{
				"id":          "c1",
				"author":      map[string]any{"displayName": "Alice"},
				"content":     "Comment one",
				"resolved":    false,
				"createdTime": "2026-02-23T00:00:00Z",
				"replies":     []map[string]any{},
			},
			{
				"id":          "c2",
				"author":      map[string]any{"displayName": "Bob"},
				"content":     "Comment two",
				"resolved":    false,
				"createdTime": "2026-02-23T01:00:00Z",
				"replies":     []map[string]any{},
			},
		},
	}
	page2 := map[string]any{
		"comments": []map[string]any{
			{
				"id":          "c3",
				"author":      map[string]any{"displayName": "Carol"},
				"content":     "Comment three",
				"resolved":    false,
				"createdTime": "2026-02-23T02:00:00Z",
				"replies":     []map[string]any{},
			},
		},
	}

	callCount := 0
	svc := newMockDriveService(t, func(req *http.Request) (*http.Response, error) {
		callCount++
		if callCount == 1 {
			return jsonResponse(200, page1)(req)
		}
		return jsonResponse(200, page2)(req)
	})

	result := listComments(context.Background(), svc, listCommentsInput{
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
	mockResp := map[string]any{
		"kind":     "drive#commentList",
		"comments": []map[string]any{},
	}
	svc := newMockDriveService(t, jsonResponse(200, mockResp))
	result := listComments(context.Background(), svc, listCommentsInput{
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
	svc := newMockDriveService(t, googleAPIError(404, "File not found."))
	result := listComments(context.Background(), svc, listCommentsInput{
		DocumentID: "nonexistent-doc",
	})
	assert.True(t, result.IsError)
	assert.NotEmpty(t, result.Content)
}
