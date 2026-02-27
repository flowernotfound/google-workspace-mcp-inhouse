package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeFilesListResponse(files []map[string]any) map[string]any {
	return map[string]any{
		"kind":  "drive#fileList",
		"files": files,
	}
}

func makeFileEntry(id, name string) map[string]any {
	return map[string]any{
		"id":           id,
		"name":         name,
		"createdTime":  "2026-02-01T00:00:00Z",
		"modifiedTime": "2026-02-23T00:00:00Z",
		"owners":       []map[string]any{{"displayName": "Alice"}},
		"webViewLink":  fmt.Sprintf("https://docs.google.com/document/d/%s/edit", id),
	}
}

func TestListDocuments_ReturnsDocuments(t *testing.T) {
	mockResp := makeFilesListResponse([]map[string]any{
		makeFileEntry("doc-1", "Document One"),
		makeFileEntry("doc-2", "Document Two"),
	})

	svc := newMockDriveService(t, jsonResponse(200, mockResp))
	result := listDocuments(context.Background(), svc, listDocumentsInput{})
	assert.False(t, result.IsError)

	text := result.Content[0].(*mcp.TextContent).Text
	var items []documentItem
	require.NoError(t, json.Unmarshal([]byte(text), &items))

	assert.Len(t, items, 2)
	assert.Equal(t, "doc-1", items[0].ID)
	assert.Equal(t, "Document One", items[0].Name)
	assert.Equal(t, []string{"Alice"}, items[0].Owners)
	assert.NotEmpty(t, items[0].WebViewLink)
}

func TestListDocuments_EmptyList(t *testing.T) {
	mockResp := makeFilesListResponse([]map[string]any{})
	svc := newMockDriveService(t, jsonResponse(200, mockResp))

	result := listDocuments(context.Background(), svc, listDocumentsInput{})
	assert.False(t, result.IsError)

	text := result.Content[0].(*mcp.TextContent).Text
	var items []documentItem
	require.NoError(t, json.Unmarshal([]byte(text), &items))
	assert.NotNil(t, items)
	assert.Empty(t, items)
}

func TestListDocuments_FolderIDFilter(t *testing.T) {
	var capturedURL string
	svc := newMockDriveService(t, func(req *http.Request) (*http.Response, error) {
		capturedURL = req.URL.RawQuery
		return jsonResponse(200, makeFilesListResponse([]map[string]any{
			makeFileEntry("doc-1", "Folder Doc"),
		}))(req)
	})

	folderID := "folder-abc"
	result := listDocuments(context.Background(), svc, listDocumentsInput{
		FolderID: &folderID,
	})
	assert.False(t, result.IsError)

	// Verify the query includes the folder filter
	assert.Contains(t, capturedURL, "folder-abc")
}

func TestListDocuments_MaxResultsClamped(t *testing.T) {
	var capturedURL string
	svc := newMockDriveService(t, func(req *http.Request) (*http.Response, error) {
		capturedURL = req.URL.RawQuery
		return jsonResponse(200, makeFilesListResponse([]map[string]any{}))(req)
	})

	over := 200
	listDocuments(context.Background(), svc, listDocumentsInput{
		MaxResults: &over,
	})

	// pageSize should be clamped to 100
	assert.Contains(t, capturedURL, "pageSize=100")
}

func TestListDocuments_DefaultMaxResults(t *testing.T) {
	var capturedURL string
	svc := newMockDriveService(t, func(req *http.Request) (*http.Response, error) {
		capturedURL = req.URL.RawQuery
		return jsonResponse(200, makeFilesListResponse([]map[string]any{}))(req)
	})

	listDocuments(context.Background(), svc, listDocumentsInput{})

	assert.Contains(t, capturedURL, fmt.Sprintf("pageSize=%d", defaultListMaxResults))
}

func TestListDocuments_CustomOrderBy(t *testing.T) {
	var capturedURL string
	svc := newMockDriveService(t, func(req *http.Request) (*http.Response, error) {
		capturedURL = req.URL.RawQuery
		return jsonResponse(200, makeFilesListResponse([]map[string]any{}))(req)
	})

	orderBy := "name"
	listDocuments(context.Background(), svc, listDocumentsInput{
		OrderBy: &orderBy,
	})

	assert.Contains(t, capturedURL, "orderBy=name")
}

func TestListDocuments_ZeroMaxResultsClampedToOne(t *testing.T) {
	var capturedURL string
	svc := newMockDriveService(t, func(req *http.Request) (*http.Response, error) {
		capturedURL = req.URL.RawQuery
		return jsonResponse(200, makeFilesListResponse([]map[string]any{}))(req)
	})

	zero := 0
	listDocuments(context.Background(), svc, listDocumentsInput{
		MaxResults: &zero,
	})

	assert.Contains(t, capturedURL, "pageSize=1")
}

func TestListDocuments_APIError(t *testing.T) {
	svc := newMockDriveService(t, googleAPIError(403, "Access denied."))

	result := listDocuments(context.Background(), svc, listDocumentsInput{})
	assert.True(t, result.IsError)

	msg := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, msg, "403")
}
