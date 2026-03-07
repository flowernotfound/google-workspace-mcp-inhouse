package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/flowernotfound/google-workspace-mcp-inhouse/internal/converter"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// readDocumentResult is the JSON response shape for read_document.
type readDocumentResult struct {
	DocumentID string `json:"document_id"`
	Title      string `json:"title"`
	Format     string `json:"format"`
	Content    string `json:"content"`
}

func readDocument(ctx context.Context, docsClient DocsClient, input readDocumentInput) *mcp.CallToolResult {
	input.DocumentID = ResolveID(input.DocumentID)
	doc, err := docsClient.GetDocument(ctx, input.DocumentID)
	if err != nil {
		return errorResult(err)
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
		}
	}

	result := readDocumentResult{
		DocumentID: doc.DocumentId,
		Title:      doc.Title,
		Format:     format,
		Content:    content,
	}
	data, err := json.Marshal(result)
	if err != nil {
		return errorResult(fmt.Errorf("failed to serialize response: %w", err))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}
}
