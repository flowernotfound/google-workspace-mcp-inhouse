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

func TestSearchSpreadsheets_ReturnsResults(t *testing.T) {
	var capturedQuery string
	mock := &mockDriveClient{
		listFilesFn: func(_ context.Context, query, _ string, _ int64, _ string) (*drive.FileList, error) {
			capturedQuery = query
			return &drive.FileList{
				Files: makeTestFiles(testFileEntry{"ss-1", "Budget 2026"}),
			}, nil
		},
	}

	result := searchSpreadsheets(context.Background(), mock, searchSpreadsheetsInput{
		Query: "Budget",
	})
	assert.False(t, result.IsError)
	assert.Contains(t, capturedQuery, sheetsFileMimeType)
	assert.Contains(t, capturedQuery, "Budget")

	text := result.Content[0].(*mcp.TextContent).Text
	var items []documentItem
	require.NoError(t, json.Unmarshal([]byte(text), &items))
	assert.Len(t, items, 1)
	assert.Equal(t, "Budget 2026", items[0].Name)
}

func TestSearchSpreadsheets_EmptyQuery(t *testing.T) {
	result := searchSpreadsheets(context.Background(), nil, searchSpreadsheetsInput{
		Query: "   ",
	})
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "empty")
}

func TestSearchSpreadsheets_SpecialCharacters(t *testing.T) {
	var capturedQuery string
	mock := &mockDriveClient{
		listFilesFn: func(_ context.Context, query, _ string, _ int64, _ string) (*drive.FileList, error) {
			capturedQuery = query
			return &drive.FileList{Files: []*drive.File{}}, nil
		},
	}

	searchSpreadsheets(context.Background(), mock, searchSpreadsheetsInput{
		Query: "it's a test",
	})
	assert.Contains(t, capturedQuery, `it\'s a test`)
}

func TestSearchSpreadsheets_APIError(t *testing.T) {
	mock := &mockDriveClient{
		listFilesFn: func(_ context.Context, _, _ string, _ int64, _ string) (*drive.FileList, error) {
			return nil, &googleapi.Error{Code: 403, Message: "access denied"}
		},
	}

	result := searchSpreadsheets(context.Background(), mock, searchSpreadsheetsInput{
		Query: "test",
	})
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "403")
}
