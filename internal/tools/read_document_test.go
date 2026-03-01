package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	docs "google.golang.org/api/docs/v1"
	"google.golang.org/api/googleapi"
)

// minimalDocument creates a minimal *docs.Document for testing.
func minimalDocument(paragraphTexts []string) *docs.Document {
	content := make([]*docs.StructuralElement, 0, len(paragraphTexts))
	for _, text := range paragraphTexts {
		content = append(content, &docs.StructuralElement{
			Paragraph: &docs.Paragraph{
				Elements: []*docs.ParagraphElement{
					{TextRun: &docs.TextRun{Content: text + "\n"}},
				},
				ParagraphStyle: &docs.ParagraphStyle{NamedStyleType: "NORMAL_TEXT"},
			},
		})
	}
	return &docs.Document{
		DocumentId: "doc-id",
		Title:      "My Document",
		Body:       &docs.Body{Content: content},
	}
}

func TestReadDocument_ReturnsMarkdown(t *testing.T) {
	mock := &mockDocsClient{
		getDocumentFn: func(_ context.Context, _ string) (*docs.Document, error) {
			return minimalDocument([]string{"Hello world"}), nil
		},
	}

	result := readDocument(context.Background(), mock, readDocumentInput{
		DocumentID: "doc-id",
	})
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
	mock := &mockDocsClient{
		getDocumentFn: func(_ context.Context, _ string) (*docs.Document, error) {
			return minimalDocument([]string{"Hello world"}), nil
		},
	}

	format := "markdown"
	result := readDocument(context.Background(), mock, readDocumentInput{
		DocumentID: "doc-id",
		Format:     &format,
	})
	assert.False(t, result.IsError)

	text := result.Content[0].(*mcp.TextContent).Text
	var res readDocumentResult
	require.NoError(t, json.Unmarshal([]byte(text), &res))
	assert.Equal(t, "markdown", res.Format)
}

func TestReadDocument_TextFormat(t *testing.T) {
	mock := &mockDocsClient{
		getDocumentFn: func(_ context.Context, _ string) (*docs.Document, error) {
			return minimalDocument([]string{"**Bold text**"}), nil
		},
	}

	format := "text"
	result := readDocument(context.Background(), mock, readDocumentInput{
		DocumentID: "doc-id",
		Format:     &format,
	})
	assert.False(t, result.IsError)

	text := result.Content[0].(*mcp.TextContent).Text
	var res readDocumentResult
	require.NoError(t, json.Unmarshal([]byte(text), &res))
	assert.Equal(t, "text", res.Format)
	// text format returns raw TextRun content as-is; "**Bold text**" is the literal mock string, not a Markdown marker
	assert.Contains(t, res.Content, "**Bold text**")
}

func TestReadDocument_InvalidFormat(t *testing.T) {
	mock := &mockDocsClient{
		getDocumentFn: func(_ context.Context, _ string) (*docs.Document, error) {
			return minimalDocument([]string{"Hello"}), nil
		},
	}

	format := "html"
	result := readDocument(context.Background(), mock, readDocumentInput{
		DocumentID: "doc-id",
		Format:     &format,
	})
	assert.True(t, result.IsError)

	msg := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, msg, "unsupported format")
}

func TestReadDocument_NotFound(t *testing.T) {
	mock := &mockDocsClient{
		getDocumentFn: func(_ context.Context, _ string) (*docs.Document, error) {
			return nil, &googleapi.Error{Code: 404, Message: "File not found."}
		},
	}

	result := readDocument(context.Background(), mock, readDocumentInput{
		DocumentID: "nonexistent-id",
	})
	assert.True(t, result.IsError)

	msg := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, msg, "404")
}

func TestReadDocument_Forbidden(t *testing.T) {
	mock := &mockDocsClient{
		getDocumentFn: func(_ context.Context, _ string) (*docs.Document, error) {
			return nil, &googleapi.Error{Code: 403, Message: "The caller does not have permission."}
		},
	}

	result := readDocument(context.Background(), mock, readDocumentInput{
		DocumentID: "doc-id",
	})
	assert.True(t, result.IsError)

	msg := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, msg, "403")
}

func TestReadDocument_AuthError(t *testing.T) {
	mock := &mockDocsClient{
		getDocumentFn: func(_ context.Context, _ string) (*docs.Document, error) {
			return nil, &googleapi.Error{Code: 401, Message: "Invalid Credentials."}
		},
	}

	result := readDocument(context.Background(), mock, readDocumentInput{
		DocumentID: "doc-id",
	})
	assert.True(t, result.IsError)

	msg := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, msg, "auth")
}

func TestReadDocument_URLInput_ExtractsID(t *testing.T) {
	const expectedID = "1A2B3C4D5E6F7G8H9I0J"

	mock := &mockDocsClient{
		getDocumentFn: func(_ context.Context, documentID string) (*docs.Document, error) {
			assert.Equal(t, expectedID, documentID)
			return minimalDocument([]string{"content"}), nil
		},
	}

	result := readDocument(context.Background(), mock, readDocumentInput{
		DocumentID: "https://docs.google.com/document/d/" + expectedID + "/edit",
	})
	assert.False(t, result.IsError)
}
