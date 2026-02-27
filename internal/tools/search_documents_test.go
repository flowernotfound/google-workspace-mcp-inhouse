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

func TestSearchDocuments_ReturnsMatches(t *testing.T) {
	mock := &mockDriveClient{
		listFilesFn: func(_ context.Context, _, _ string, _ int64, _ string) (*drive.FileList, error) {
			return &drive.FileList{
				Files: makeTestFiles(
					struct{ id, name string }{"doc-1", "Meeting Notes"},
					struct{ id, name string }{"doc-2", "Meeting Agenda"},
				),
			}, nil
		},
	}

	result := searchDocuments(context.Background(), mock, searchDocumentsInput{
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
	mock := &mockDriveClient{
		listFilesFn: func(_ context.Context, _, _ string, _ int64, _ string) (*drive.FileList, error) {
			return &drive.FileList{Files: []*drive.File{}}, nil
		},
	}

	result := searchDocuments(context.Background(), mock, searchDocumentsInput{
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
	var capturedQuery string
	mock := &mockDriveClient{
		listFilesFn: func(_ context.Context, query, _ string, _ int64, _ string) (*drive.FileList, error) {
			capturedQuery = query
			return &drive.FileList{Files: []*drive.File{}}, nil
		},
	}

	searchDocuments(context.Background(), mock, searchDocumentsInput{
		Query: "it's a test",
	})

	// Single quote in query must be escaped as \'
	assert.Contains(t, capturedQuery, `it\'s`)
}

func TestSearchDocuments_MaxResultsClamped(t *testing.T) {
	var capturedPageSize int64
	mock := &mockDriveClient{
		listFilesFn: func(_ context.Context, _, _ string, pageSize int64, _ string) (*drive.FileList, error) {
			capturedPageSize = pageSize
			return &drive.FileList{Files: []*drive.File{}}, nil
		},
	}

	over := 200
	searchDocuments(context.Background(), mock, searchDocumentsInput{
		Query:      "test",
		MaxResults: &over,
	})

	assert.Equal(t, int64(50), capturedPageSize)
}

func TestSearchDocuments_DefaultMaxResults(t *testing.T) {
	var capturedPageSize int64
	mock := &mockDriveClient{
		listFilesFn: func(_ context.Context, _, _ string, pageSize int64, _ string) (*drive.FileList, error) {
			capturedPageSize = pageSize
			return &drive.FileList{Files: []*drive.File{}}, nil
		},
	}

	searchDocuments(context.Background(), mock, searchDocumentsInput{
		Query: "test",
	})

	assert.Equal(t, int64(10), capturedPageSize)
}

func TestSearchDocuments_ZeroMaxResultsClampedToOne(t *testing.T) {
	var capturedPageSize int64
	mock := &mockDriveClient{
		listFilesFn: func(_ context.Context, _, _ string, pageSize int64, _ string) (*drive.FileList, error) {
			capturedPageSize = pageSize
			return &drive.FileList{Files: []*drive.File{}}, nil
		},
	}

	zero := 0
	searchDocuments(context.Background(), mock, searchDocumentsInput{
		Query:      "test",
		MaxResults: &zero,
	})

	assert.Equal(t, int64(1), capturedPageSize)
}

func TestSearchDocuments_EmptyQuery(t *testing.T) {
	mock := &mockDriveClient{
		listFilesFn: func(_ context.Context, _, _ string, _ int64, _ string) (*drive.FileList, error) {
			return &drive.FileList{Files: []*drive.File{}}, nil
		},
	}

	result := searchDocuments(context.Background(), mock, searchDocumentsInput{
		Query: "",
	})
	assert.True(t, result.IsError)

	msg := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, msg, "must not be empty")
}

func TestSearchDocuments_WhitespaceQuery(t *testing.T) {
	mock := &mockDriveClient{
		listFilesFn: func(_ context.Context, _, _ string, _ int64, _ string) (*drive.FileList, error) {
			return &drive.FileList{Files: []*drive.File{}}, nil
		},
	}

	result := searchDocuments(context.Background(), mock, searchDocumentsInput{
		Query: "   ",
	})
	assert.True(t, result.IsError)

	msg := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, msg, "must not be empty")
}

func TestSearchDocuments_APIError(t *testing.T) {
	mock := &mockDriveClient{
		listFilesFn: func(_ context.Context, _, _ string, _ int64, _ string) (*drive.FileList, error) {
			return nil, &googleapi.Error{Code: 403, Message: "Access denied."}
		},
	}

	result := searchDocuments(context.Background(), mock, searchDocumentsInput{
		Query: "test",
	})
	assert.True(t, result.IsError)

	msg := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, msg, "403")
}
