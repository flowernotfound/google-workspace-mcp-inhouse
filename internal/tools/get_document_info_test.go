package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeFileGetResponse(id, name, description string) map[string]any {
	return map[string]any{
		"kind":         "drive#file",
		"id":           id,
		"name":         name,
		"description":  description,
		"createdTime":  "2026-01-01T00:00:00Z",
		"modifiedTime": "2026-02-23T12:00:00Z",
		"owners": []map[string]any{
			{"displayName": "Alice"},
		},
		"lastModifyingUser": map[string]any{
			"displayName": "Bob",
		},
		"webViewLink": "https://docs.google.com/document/d/" + id + "/edit",
	}
}

func TestGetDocumentInfo_ReturnsFullInfo(t *testing.T) {
	mockResp := makeFileGetResponse("doc-id", "My Document", "A test document")
	svc := newMockDriveService(t, jsonResponse(200, mockResp))

	result, _, err := getDocumentInfo(context.Background(), svc, getDocumentInfoInput{
		DocumentID: "doc-id",
	})
	require.NoError(t, err)
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
	mockResp := makeFileGetResponse("doc-id", "No Description", "")
	svc := newMockDriveService(t, jsonResponse(200, mockResp))

	result, _, err := getDocumentInfo(context.Background(), svc, getDocumentInfoInput{
		DocumentID: "doc-id",
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)

	text := result.Content[0].(*mcp.TextContent).Text
	var info documentInfo
	require.NoError(t, json.Unmarshal([]byte(text), &info))

	assert.Equal(t, "", info.Description)
}

func TestGetDocumentInfo_NoLastModifyingUser(t *testing.T) {
	mockResp := map[string]any{
		"kind":         "drive#file",
		"id":           "doc-id",
		"name":         "My Document",
		"description":  "",
		"createdTime":  "2026-01-01T00:00:00Z",
		"modifiedTime": "2026-02-23T12:00:00Z",
		"owners":       []map[string]any{{"displayName": "Alice"}},
		// lastModifyingUser is absent
		"webViewLink": "https://docs.google.com/document/d/doc-id/edit",
	}
	svc := newMockDriveService(t, jsonResponse(200, mockResp))

	result, _, err := getDocumentInfo(context.Background(), svc, getDocumentInfoInput{
		DocumentID: "doc-id",
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)

	text := result.Content[0].(*mcp.TextContent).Text
	var info documentInfo
	require.NoError(t, json.Unmarshal([]byte(text), &info))

	assert.Equal(t, "", info.LastModifyingUser)
}

func TestGetDocumentInfo_NoOwners(t *testing.T) {
	mockResp := map[string]any{
		"kind":         "drive#file",
		"id":           "doc-id",
		"name":         "Shared Document",
		"description":  "",
		"createdTime":  "2026-01-01T00:00:00Z",
		"modifiedTime": "2026-02-23T12:00:00Z",
		"owners":       []map[string]any{},
		"webViewLink":  "https://docs.google.com/document/d/doc-id/edit",
	}
	svc := newMockDriveService(t, jsonResponse(200, mockResp))

	result, _, err := getDocumentInfo(context.Background(), svc, getDocumentInfoInput{
		DocumentID: "doc-id",
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)

	text := result.Content[0].(*mcp.TextContent).Text
	var info documentInfo
	require.NoError(t, json.Unmarshal([]byte(text), &info))

	assert.NotNil(t, info.Owners)
	assert.Empty(t, info.Owners)
}

func TestGetDocumentInfo_NotFound(t *testing.T) {
	svc := newMockDriveService(t, googleAPIError(404, "File not found."))

	result, _, err := getDocumentInfo(context.Background(), svc, getDocumentInfoInput{
		DocumentID: "nonexistent-id",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)

	msg := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, msg, "404")
}

func TestGetDocumentInfo_Forbidden(t *testing.T) {
	svc := newMockDriveService(t, googleAPIError(403, "Access denied."))

	result, _, err := getDocumentInfo(context.Background(), svc, getDocumentInfoInput{
		DocumentID: "doc-id",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)

	msg := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, msg, "403")
}

func TestGetDocumentInfo_AuthError(t *testing.T) {
	svc := newMockDriveService(t, googleAPIError(401, "Invalid Credentials."))

	result, _, err := getDocumentInfo(context.Background(), svc, getDocumentInfoInput{
		DocumentID: "doc-id",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)

	msg := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, msg, "auth")
}
