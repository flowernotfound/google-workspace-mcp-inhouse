package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchDocuments_ReturnsMatches(t *testing.T) {
	mockResp := makeFilesListResponse([]map[string]any{
		makeFileEntry("doc-1", "Meeting Notes"),
		makeFileEntry("doc-2", "Meeting Agenda"),
	})

	svc := newMockDriveService(t, jsonResponse(200, mockResp))
	result := searchDocuments(context.Background(), svc, searchDocumentsInput{
		Query: "meeting",
	})
	assert.False(t, result.IsError)

	text := result.Content[0].(*mcp.TextContent).Text
	var items []documentItem
	require.NoError(t, json.Unmarshal([]byte(text), &items))

	assert.Len(t, items, 2)
	assert.Equal(t, "doc-1", items[0].ID)
	assert.Equal(t, "Meeting Notes", items[0].Name)
}

func TestSearchDocuments_EmptyResults(t *testing.T) {
	mockResp := makeFilesListResponse([]map[string]any{})
	svc := newMockDriveService(t, jsonResponse(200, mockResp))

	result := searchDocuments(context.Background(), svc, searchDocumentsInput{
		Query: "nonexistent keyword xyz",
	})
	assert.False(t, result.IsError)

	text := result.Content[0].(*mcp.TextContent).Text
	var items []documentItem
	require.NoError(t, json.Unmarshal([]byte(text), &items))
	assert.NotNil(t, items)
	assert.Empty(t, items)
}

func TestSearchDocuments_SingleQuoteEscaped(t *testing.T) {
	var capturedURL string
	svc := newMockDriveService(t, func(req *http.Request) (*http.Response, error) {
		capturedURL = req.URL.RawQuery
		return jsonResponse(200, makeFilesListResponse([]map[string]any{}))(req)
	})

	searchDocuments(context.Background(), svc, searchDocumentsInput{
		Query: "it's a test",
	})

	// Single quote in query must be escaped as \' which URL-encodes to %5C%27
	assert.Contains(t, capturedURL, `it%5C%27s`)
}

func TestSearchDocuments_MaxResultsClamped(t *testing.T) {
	var capturedURL string
	svc := newMockDriveService(t, func(req *http.Request) (*http.Response, error) {
		capturedURL = req.URL.RawQuery
		return jsonResponse(200, makeFilesListResponse([]map[string]any{}))(req)
	})

	over := 200
	searchDocuments(context.Background(), svc, searchDocumentsInput{
		Query:      "test",
		MaxResults: &over,
	})

	assert.Contains(t, capturedURL, "pageSize=50")
}

func TestSearchDocuments_DefaultMaxResults(t *testing.T) {
	var capturedURL string
	svc := newMockDriveService(t, func(req *http.Request) (*http.Response, error) {
		capturedURL = req.URL.RawQuery
		return jsonResponse(200, makeFilesListResponse([]map[string]any{}))(req)
	})

	searchDocuments(context.Background(), svc, searchDocumentsInput{
		Query: "test",
	})

	assert.Contains(t, capturedURL, "pageSize=10")
}

func TestSearchDocuments_ZeroMaxResultsClampedToOne(t *testing.T) {
	var capturedURL string
	svc := newMockDriveService(t, func(req *http.Request) (*http.Response, error) {
		capturedURL = req.URL.RawQuery
		return jsonResponse(200, makeFilesListResponse([]map[string]any{}))(req)
	})

	zero := 0
	searchDocuments(context.Background(), svc, searchDocumentsInput{
		Query:      "test",
		MaxResults: &zero,
	})

	assert.Contains(t, capturedURL, "pageSize=1")
}

func TestSearchDocuments_EmptyQuery(t *testing.T) {
	svc := newMockDriveService(t, jsonResponse(200, makeFilesListResponse([]map[string]any{})))

	result := searchDocuments(context.Background(), svc, searchDocumentsInput{
		Query: "",
	})
	assert.True(t, result.IsError)

	msg := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, msg, "must not be empty")
}

func TestSearchDocuments_WhitespaceQuery(t *testing.T) {
	svc := newMockDriveService(t, jsonResponse(200, makeFilesListResponse([]map[string]any{})))

	result := searchDocuments(context.Background(), svc, searchDocumentsInput{
		Query: "   ",
	})
	assert.True(t, result.IsError)

	msg := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, msg, "must not be empty")
}

func TestSearchDocuments_APIError(t *testing.T) {
	svc := newMockDriveService(t, googleAPIError(403, "Access denied."))

	result := searchDocuments(context.Background(), svc, searchDocumentsInput{
		Query: "test",
	})
	assert.True(t, result.IsError)

	msg := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, msg, "403")
}
