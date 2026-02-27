package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	docs "google.golang.org/api/docs/v1"
	drive "google.golang.org/api/drive/v3"
)

// Input types for each tool.

type readDocumentInput struct {
	DocumentID string  `json:"document_id" jsonschema:"required,the ID of the Google Docs document"`
	Format     *string `json:"format,omitempty" jsonschema:"output format: 'markdown' (default) or 'text'"`
}

type listDocumentsInput struct {
	FolderID   *string `json:"folder_id,omitempty" jsonschema:"folder ID to filter documents"`
	MaxResults *int    `json:"max_results,omitempty" jsonschema:"maximum number of results (default 20, max 100)"`
	OrderBy    *string `json:"order_by,omitempty" jsonschema:"sort order, e.g. 'modifiedTime desc' (default), 'name', 'createdTime'"`
}

type searchDocumentsInput struct {
	Query      string `json:"query" jsonschema:"search keyword"`
	MaxResults *int   `json:"max_results,omitempty" jsonschema:"maximum number of results (default 10, max 50)"`
}

type getDocumentInfoInput struct {
	DocumentID string `json:"document_id" jsonschema:"required,the ID of the Google Docs document"`
}

type listCommentsInput struct {
	DocumentID      string `json:"document_id" jsonschema:"required,the ID of the Google Docs document"`
	IncludeResolved bool   `json:"include_resolved" jsonschema:"whether to include resolved comments (default false)"`
}

type getCommentInput struct {
	DocumentID string `json:"document_id" jsonschema:"required,the ID of the Google Docs document"`
	CommentID  string `json:"comment_id" jsonschema:"required,the comment ID"`
}

// RegisterTools registers all MCP tools on the server.
func RegisterTools(server *mcp.Server, docsService *docs.Service, driveService *drive.Service) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "read_document",
		Description: "Google Docs ドキュメントの本文を Markdown 形式またはプレーンテキストで取得する",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input readDocumentInput) (*mcp.CallToolResult, any, error) {
		return readDocument(ctx, docsService, input), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_documents",
		Description: "Google Drive 内の Google Docs ドキュメント一覧を取得する",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input listDocumentsInput) (*mcp.CallToolResult, any, error) {
		return listDocuments(ctx, driveService, input), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_documents",
		Description: "キーワードで Google Docs ドキュメントを検索する",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input searchDocumentsInput) (*mcp.CallToolResult, any, error) {
		return searchDocuments(ctx, driveService, input), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_document_info",
		Description: "Google Docs ドキュメントのメタ情報（タイトル、作成日、更新日、オーナー等）を取得する",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input getDocumentInfoInput) (*mcp.CallToolResult, any, error) {
		return getDocumentInfo(ctx, driveService, input), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_comments",
		Description: "Google Docs ドキュメントのコメント一覧を取得する",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input listCommentsInput) (*mcp.CallToolResult, any, error) {
		return listComments(ctx, driveService, input), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_comment",
		Description: "Google Docs ドキュメントの個別コメントと返信スレッドを取得する",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input getCommentInput) (*mcp.CallToolResult, any, error) {
		return getComment(ctx, driveService, input), nil, nil
	})
}
