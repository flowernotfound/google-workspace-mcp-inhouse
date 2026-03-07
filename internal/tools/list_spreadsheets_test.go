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

func TestListSpreadsheets_ReturnsSpreadsheets(t *testing.T) {
	mock := &mockDriveClient{
		listFilesFn: func(_ context.Context, query, _ string, _ int64, _ string) (*drive.FileList, error) {
			assert.Contains(t, query, sheetsFileMimeType)
			return &drive.FileList{
				Files: makeTestFiles(
					testFileEntry{"ss-1", "Spreadsheet One"},
					testFileEntry{"ss-2", "Spreadsheet Two"},
				),
			}, nil
		},
	}

	result := listSpreadsheets(context.Background(), mock, listSpreadsheetsInput{})
	assert.False(t, result.IsError)

	text := result.Content[0].(*mcp.TextContent).Text
	var items []documentItem
	require.NoError(t, json.Unmarshal([]byte(text), &items))

	assert.Len(t, items, 2)
	assert.Equal(t, "ss-1", items[0].ID)
	assert.Equal(t, "Spreadsheet One", items[0].Name)
}

func TestListSpreadsheets_EmptyList(t *testing.T) {
	mock := &mockDriveClient{
		listFilesFn: func(_ context.Context, _, _ string, _ int64, _ string) (*drive.FileList, error) {
			return &drive.FileList{Files: []*drive.File{}}, nil
		},
	}

	result := listSpreadsheets(context.Background(), mock, listSpreadsheetsInput{})
	assert.False(t, result.IsError)

	text := result.Content[0].(*mcp.TextContent).Text
	var items []documentItem
	require.NoError(t, json.Unmarshal([]byte(text), &items))
	assert.Empty(t, items)
}

func TestListSpreadsheets_MaxResultsClamped(t *testing.T) {
	var capturedPageSize int64
	mock := &mockDriveClient{
		listFilesFn: func(_ context.Context, _, _ string, pageSize int64, _ string) (*drive.FileList, error) {
			capturedPageSize = pageSize
			return &drive.FileList{Files: []*drive.File{}}, nil
		},
	}

	over := 200
	listSpreadsheets(context.Background(), mock, listSpreadsheetsInput{
		MaxResults: &over,
	})
	assert.Equal(t, int64(100), capturedPageSize)
}

func TestListSpreadsheets_FolderIDFilter(t *testing.T) {
	var capturedQuery string
	mock := &mockDriveClient{
		listFilesFn: func(_ context.Context, query, _ string, _ int64, _ string) (*drive.FileList, error) {
			capturedQuery = query
			return &drive.FileList{Files: []*drive.File{}}, nil
		},
	}

	folderID := "folder-xyz"
	listSpreadsheets(context.Background(), mock, listSpreadsheetsInput{
		FolderID: &folderID,
	})
	assert.Contains(t, capturedQuery, "folder-xyz")
	assert.Contains(t, capturedQuery, sheetsFileMimeType)
}

func TestListSpreadsheets_APIError(t *testing.T) {
	mock := &mockDriveClient{
		listFilesFn: func(_ context.Context, _, _ string, _ int64, _ string) (*drive.FileList, error) {
			return nil, &googleapi.Error{Code: 403, Message: "access denied"}
		},
	}

	result := listSpreadsheets(context.Background(), mock, listSpreadsheetsInput{})
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "403")
}
