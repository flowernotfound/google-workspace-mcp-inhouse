package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	drive "google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
)

func makeTestFiles(entries ...struct{ id, name string }) []*drive.File {
	files := make([]*drive.File, 0, len(entries))
	for _, e := range entries {
		files = append(files, &drive.File{
			Id:           e.id,
			Name:         e.name,
			CreatedTime:  "2026-02-01T00:00:00Z",
			ModifiedTime: "2026-02-23T00:00:00Z",
			Owners:       []*drive.User{{DisplayName: "Alice"}},
			WebViewLink:  fmt.Sprintf("https://docs.google.com/document/d/%s/edit", e.id),
		})
	}
	return files
}

func TestListDocuments_ReturnsDocuments(t *testing.T) {
	mock := &mockDriveClient{
		listFilesFn: func(_ context.Context, _, _ string, _ int64, _ string) (*drive.FileList, error) {
			return &drive.FileList{
				Files: makeTestFiles(
					struct{ id, name string }{"doc-1", "Document One"},
					struct{ id, name string }{"doc-2", "Document Two"},
				),
			}, nil
		},
	}

	result := listDocuments(context.Background(), mock, listDocumentsInput{})
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
	mock := &mockDriveClient{
		listFilesFn: func(_ context.Context, _, _ string, _ int64, _ string) (*drive.FileList, error) {
			return &drive.FileList{Files: []*drive.File{}}, nil
		},
	}

	result := listDocuments(context.Background(), mock, listDocumentsInput{})
	assert.False(t, result.IsError)

	text := result.Content[0].(*mcp.TextContent).Text
	var items []documentItem
	require.NoError(t, json.Unmarshal([]byte(text), &items))
	assert.NotNil(t, items)
	assert.Empty(t, items)
}

func TestListDocuments_FolderIDFilter(t *testing.T) {
	var capturedQuery string
	mock := &mockDriveClient{
		listFilesFn: func(_ context.Context, query, _ string, _ int64, _ string) (*drive.FileList, error) {
			capturedQuery = query
			return &drive.FileList{
				Files: makeTestFiles(struct{ id, name string }{"doc-1", "Folder Doc"}),
			}, nil
		},
	}

	folderID := "folder-abc"
	result := listDocuments(context.Background(), mock, listDocumentsInput{
		FolderID: &folderID,
	})
	assert.False(t, result.IsError)

	// Verify the query includes the folder filter
	assert.Contains(t, capturedQuery, "folder-abc")
}

func TestListDocuments_MaxResultsClamped(t *testing.T) {
	var capturedPageSize int64
	mock := &mockDriveClient{
		listFilesFn: func(_ context.Context, _, _ string, pageSize int64, _ string) (*drive.FileList, error) {
			capturedPageSize = pageSize
			return &drive.FileList{Files: []*drive.File{}}, nil
		},
	}

	over := 200
	listDocuments(context.Background(), mock, listDocumentsInput{
		MaxResults: &over,
	})

	// pageSize should be clamped to 100
	assert.Equal(t, int64(100), capturedPageSize)
}

func TestListDocuments_DefaultMaxResults(t *testing.T) {
	var capturedPageSize int64
	mock := &mockDriveClient{
		listFilesFn: func(_ context.Context, _, _ string, pageSize int64, _ string) (*drive.FileList, error) {
			capturedPageSize = pageSize
			return &drive.FileList{Files: []*drive.File{}}, nil
		},
	}

	listDocuments(context.Background(), mock, listDocumentsInput{})

	assert.Equal(t, int64(defaultListMaxResults), capturedPageSize)
}

func TestListDocuments_CustomOrderBy(t *testing.T) {
	var capturedOrderBy string
	mock := &mockDriveClient{
		listFilesFn: func(_ context.Context, _, orderBy string, _ int64, _ string) (*drive.FileList, error) {
			capturedOrderBy = orderBy
			return &drive.FileList{Files: []*drive.File{}}, nil
		},
	}

	orderBy := "name"
	listDocuments(context.Background(), mock, listDocumentsInput{
		OrderBy: &orderBy,
	})

	assert.Equal(t, "name", capturedOrderBy)
}

func TestListDocuments_ZeroMaxResultsClampedToOne(t *testing.T) {
	var capturedPageSize int64
	mock := &mockDriveClient{
		listFilesFn: func(_ context.Context, _, _ string, pageSize int64, _ string) (*drive.FileList, error) {
			capturedPageSize = pageSize
			return &drive.FileList{Files: []*drive.File{}}, nil
		},
	}

	zero := 0
	listDocuments(context.Background(), mock, listDocumentsInput{
		MaxResults: &zero,
	})

	assert.Equal(t, int64(1), capturedPageSize)
}

func TestListDocuments_APIError(t *testing.T) {
	mock := &mockDriveClient{
		listFilesFn: func(_ context.Context, _, _ string, _ int64, _ string) (*drive.FileList, error) {
			return nil, &googleapi.Error{Code: 403, Message: "Access denied."}
		},
	}

	result := listDocuments(context.Background(), mock, listDocumentsInput{})
	assert.True(t, result.IsError)

	msg := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, msg, "403")
}
