package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	docs "google.golang.org/api/docs/v1"
	"google.golang.org/api/option"
)

// newMockDocsService creates a Docs service backed by a mock transport.
func newMockDocsService(t *testing.T, fn func(*http.Request) (*http.Response, error)) *docs.Service {
	t.Helper()
	svc, err := docs.NewService(context.Background(),
		option.WithHTTPClient(&http.Client{Transport: &mockTransport{fn: fn}}),
	)
	require.NoError(t, err)
	return svc
}

// minimalDocsResponse returns a minimal Google Docs API response JSON body.
func minimalDocsResponse(documentID, title string, paragraphTexts []string) map[string]any {
	content := make([]map[string]any, 0, len(paragraphTexts))
	for _, text := range paragraphTexts {
		content = append(content, map[string]any{
			"paragraph": map[string]any{
				"elements": []map[string]any{
					{
						"textRun": map[string]any{
							"content": text + "\n",
						},
					},
				},
				"paragraphStyle": map[string]any{
					"namedStyleType": "NORMAL_TEXT",
				},
			},
		})
	}
	return map[string]any{
		"documentId": documentID,
		"title":      title,
		"body": map[string]any{
			"content": content,
		},
	}
}

func TestReadDocument_ReturnsMarkdown(t *testing.T) {
	mockResp := minimalDocsResponse("doc-id", "My Document", []string{"Hello world"})
	svc := newMockDocsService(t, jsonResponse(200, mockResp))

	result, _, err := readDocument(context.Background(), svc, readDocumentInput{
		DocumentID: "doc-id",
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)

	text := result.Content[0].(*mcp.TextContent).Text
	var res readDocumentResult
	require.NoError(t, json.Unmarshal([]byte(text), &res))

	assert.Equal(t, "doc-id", res.DocumentID)
	assert.Equal(t, "My Document", res.Title)
	assert.Equal(t, "markdown", res.Format)
	assert.Contains(t, res.Content, "Hello world")
}

func TestReadDocument_ExplicitMarkdownFormat(t *testing.T) {
	mockResp := minimalDocsResponse("doc-id", "My Document", []string{"Hello world"})
	svc := newMockDocsService(t, jsonResponse(200, mockResp))

	format := "markdown"
	result, _, err := readDocument(context.Background(), svc, readDocumentInput{
		DocumentID: "doc-id",
		Format:     &format,
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)

	text := result.Content[0].(*mcp.TextContent).Text
	var res readDocumentResult
	require.NoError(t, json.Unmarshal([]byte(text), &res))
	assert.Equal(t, "markdown", res.Format)
}

func TestReadDocument_TextFormat(t *testing.T) {
	mockResp := minimalDocsResponse("doc-id", "My Document", []string{"**Bold text**"})
	svc := newMockDocsService(t, jsonResponse(200, mockResp))

	format := "text"
	result, _, err := readDocument(context.Background(), svc, readDocumentInput{
		DocumentID: "doc-id",
		Format:     &format,
	})
	require.NoError(t, err)
	assert.False(t, result.IsError)

	text := result.Content[0].(*mcp.TextContent).Text
	var res readDocumentResult
	require.NoError(t, json.Unmarshal([]byte(text), &res))
	assert.Equal(t, "text", res.Format)
	// Text format returns raw content without Markdown markers
	assert.Contains(t, res.Content, "**Bold text**")
	assert.Equal(t, "text", res.Format)
}

func TestReadDocument_InvalidFormat(t *testing.T) {
	mockResp := minimalDocsResponse("doc-id", "My Document", []string{"Hello"})
	svc := newMockDocsService(t, jsonResponse(200, mockResp))

	format := "html"
	result, _, err := readDocument(context.Background(), svc, readDocumentInput{
		DocumentID: "doc-id",
		Format:     &format,
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)

	msg := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, msg, "unsupported format")
}

func TestReadDocument_NotFound(t *testing.T) {
	svc := newMockDocsService(t, googleAPIError(404, "File not found."))

	result, _, err := readDocument(context.Background(), svc, readDocumentInput{
		DocumentID: "nonexistent-id",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)

	msg := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, msg, "404")
}

func TestReadDocument_Forbidden(t *testing.T) {
	svc := newMockDocsService(t, googleAPIError(403, "The caller does not have permission."))

	result, _, err := readDocument(context.Background(), svc, readDocumentInput{
		DocumentID: "doc-id",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)

	msg := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, msg, "403")
}

func TestReadDocument_AuthError(t *testing.T) {
	svc := newMockDocsService(t, googleAPIError(401, "Invalid Credentials."))

	result, _, err := readDocument(context.Background(), svc, readDocumentInput{
		DocumentID: "doc-id",
	})
	require.NoError(t, err)
	assert.True(t, result.IsError)

	msg := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, msg, "auth")
}
