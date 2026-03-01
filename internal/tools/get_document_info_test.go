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

func TestGetDocumentInfo_ReturnsFullInfo(t *testing.T) {
	mock := &mockDriveClient{
		getFileFn: func(_ context.Context, fileID, _ string) (*drive.File, error) {
			return &drive.File{
				Id:                fileID,
				Name:              "My Document",
				Description:       "A test document",
				CreatedTime:       "2026-01-01T00:00:00Z",
				ModifiedTime:      "2026-02-23T12:00:00Z",
				Owners:            []*drive.User{{DisplayName: "Alice"}},
				LastModifyingUser: &drive.User{DisplayName: "Bob"},
				WebViewLink:       "https://docs.google.com/document/d/doc-id/edit",
			}, nil
		},
	}

	result := getDocumentInfo(context.Background(), mock, getDocumentInfoInput{
		DocumentID: "doc-id",
	})
	assert.False(t, result.IsError)

	text := result.Content[0].(*mcp.TextContent).Text
	var info documentInfo
	require.NoError(t, json.Unmarshal([]byte(text), &info))

	assert.Equal(t, "doc-id", info.ID)
	assert.Equal(t, "My Document", info.Name)
	assert.Equal(t, "A test document", info.Description)
	assert.Equal(t, "2026-01-01T00:00:00Z", info.CreatedTime)
	assert.Equal(t, "2026-02-23T12:00:00Z", info.ModifiedTime)
	assert.Equal(t, []string{"Alice"}, info.Owners)
	assert.Equal(t, "Bob", info.LastModifyingUser)
	assert.Contains(t, info.WebViewLink, "doc-id")
}

func TestGetDocumentInfo_EmptyDescription(t *testing.T) {
	mock := &mockDriveClient{
		getFileFn: func(_ context.Context, fileID, _ string) (*drive.File, error) {
			return &drive.File{
				Id:                fileID,
				Name:              "No Description",
				Description:       "",
				CreatedTime:       "2026-01-01T00:00:00Z",
				ModifiedTime:      "2026-02-23T12:00:00Z",
				Owners:            []*drive.User{{DisplayName: "Alice"}},
				LastModifyingUser: &drive.User{DisplayName: "Bob"},
				WebViewLink:       "https://docs.google.com/document/d/doc-id/edit",
			}, nil
		},
	}

	result := getDocumentInfo(context.Background(), mock, getDocumentInfoInput{
		DocumentID: "doc-id",
	})
	assert.False(t, result.IsError)

	text := result.Content[0].(*mcp.TextContent).Text
	var info documentInfo
	require.NoError(t, json.Unmarshal([]byte(text), &info))

	assert.Equal(t, "", info.Description)
}

func TestGetDocumentInfo_NoLastModifyingUser(t *testing.T) {
	mock := &mockDriveClient{
		getFileFn: func(_ context.Context, fileID, _ string) (*drive.File, error) {
			return &drive.File{
				Id:           fileID,
				Name:         "My Document",
				CreatedTime:  "2026-01-01T00:00:00Z",
				ModifiedTime: "2026-02-23T12:00:00Z",
				Owners:       []*drive.User{{DisplayName: "Alice"}},
				WebViewLink:  "https://docs.google.com/document/d/doc-id/edit",
			}, nil
		},
	}

	result := getDocumentInfo(context.Background(), mock, getDocumentInfoInput{
		DocumentID: "doc-id",
	})
	assert.False(t, result.IsError)

	text := result.Content[0].(*mcp.TextContent).Text
	var info documentInfo
	require.NoError(t, json.Unmarshal([]byte(text), &info))

	assert.Equal(t, "", info.LastModifyingUser)
}

func TestGetDocumentInfo_NoOwners(t *testing.T) {
	mock := &mockDriveClient{
		getFileFn: func(_ context.Context, fileID, _ string) (*drive.File, error) {
			return &drive.File{
				Id:           fileID,
				Name:         "Shared Document",
				CreatedTime:  "2026-01-01T00:00:00Z",
				ModifiedTime: "2026-02-23T12:00:00Z",
				Owners:       []*drive.User{},
				WebViewLink:  "https://docs.google.com/document/d/doc-id/edit",
			}, nil
		},
	}

	result := getDocumentInfo(context.Background(), mock, getDocumentInfoInput{
		DocumentID: "doc-id",
	})
	assert.False(t, result.IsError)

	text := result.Content[0].(*mcp.TextContent).Text
	var info documentInfo
	require.NoError(t, json.Unmarshal([]byte(text), &info))

	assert.NotNil(t, info.Owners)
	assert.Empty(t, info.Owners)
}

func TestGetDocumentInfo_NotFound(t *testing.T) {
	mock := &mockDriveClient{
		getFileFn: func(_ context.Context, _, _ string) (*drive.File, error) {
			return nil, &googleapi.Error{Code: 404, Message: "File not found."}
		},
	}

	result := getDocumentInfo(context.Background(), mock, getDocumentInfoInput{
		DocumentID: "nonexistent-id",
	})
	assert.True(t, result.IsError)

	msg := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, msg, "404")
}

func TestGetDocumentInfo_Forbidden(t *testing.T) {
	mock := &mockDriveClient{
		getFileFn: func(_ context.Context, _, _ string) (*drive.File, error) {
			return nil, &googleapi.Error{Code: 403, Message: "Access denied."}
		},
	}

	result := getDocumentInfo(context.Background(), mock, getDocumentInfoInput{
		DocumentID: "doc-id",
	})
	assert.True(t, result.IsError)

	msg := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, msg, "403")
}

func TestGetDocumentInfo_AuthError(t *testing.T) {
	mock := &mockDriveClient{
		getFileFn: func(_ context.Context, _, _ string) (*drive.File, error) {
			return nil, &googleapi.Error{Code: 401, Message: "Invalid Credentials."}
		},
	}

	result := getDocumentInfo(context.Background(), mock, getDocumentInfoInput{
		DocumentID: "doc-id",
	})
	assert.True(t, result.IsError)

	msg := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, msg, "auth")
}

func TestGetDocumentInfo_URLInput_ExtractsID(t *testing.T) {
	const expectedID = "1A2B3C4D5E6F7G8H9I0J"

	mock := &mockDriveClient{
		getFileFn: func(_ context.Context, fileID, _ string) (*drive.File, error) {
			assert.Equal(t, expectedID, fileID)
			return &drive.File{
				Id:           fileID,
				Name:         "Test",
				CreatedTime:  "2026-01-01T00:00:00Z",
				ModifiedTime: "2026-01-01T00:00:00Z",
				Owners:       []*drive.User{},
			}, nil
		},
	}

	result := getDocumentInfo(context.Background(), mock, getDocumentInfoInput{
		DocumentID: "https://docs.google.com/document/d/" + expectedID + "/edit",
	})
	assert.False(t, result.IsError)
}
