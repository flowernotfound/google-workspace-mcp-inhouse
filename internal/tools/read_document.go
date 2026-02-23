package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/flowernotfound/google-workspace-mcp-inhouse/internal/converter"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	docs "google.golang.org/api/docs/v1"
)

// readDocumentResult is the JSON response shape for read_document.
type readDocumentResult struct {
	DocumentID string `json:"document_id"`
	Title      string `json:"title"`
	Format     string `json:"format"`
	Content    string `json:"content"`
}

func readDocument(ctx context.Context, docsService *docs.Service, input readDocumentInput) (*mcp.CallToolResult, struct{}, error) {
	doc, err := docsService.Documents.Get(input.DocumentID).Context(ctx).Do()
	if err != nil {
		return errorResult(err), struct{}{}, nil
	}

	format := "markdown"
	if input.Format != nil {
		format = *input.Format
	}

	var content string
	switch format {
	case "markdown":
		content = converter.ConvertDocsToMarkdown(doc)
	case "text":
		content = converter.ConvertDocsToPlainText(doc)
	default:
		msg := fmt.Sprintf("unsupported format %q: use 'markdown' or 'text'", format)
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{Text: msg}},
		}, struct{}{}, nil
	}

	result := readDocumentResult{
		DocumentID: doc.DocumentId,
		Title:      doc.Title,
		Format:     format,
		Content:    content,
	}
	data, err := json.Marshal(result)
	if err != nil {
		return errorResult(fmt.Errorf("failed to serialize response: %w", err)), struct{}{}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, struct{}{}, nil
}
