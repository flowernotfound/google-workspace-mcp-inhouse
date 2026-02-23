package tools

import (
	"context"
	"errors"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	docs "google.golang.org/api/docs/v1"
	drive "google.golang.org/api/drive/v3"
)

// Input types for each tool.

type readDocumentInput struct {
	DocumentID string  `json:"document_id" jsonschema:"the ID of the Google Docs document"`
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

// errNotImplemented is returned by stub tool handlers until PR5 implementations land.
var errNotImplemented = errors.New("not yet implemented")

// RegisterTools registers all MCP tools on the server.
// Document tool implementations (read_document, list_documents, search_documents, get_document_info)
// will be added in PR5.
func RegisterTools(server *mcp.Server, docsService *docs.Service, driveService *drive.Service) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "read_document",
		Description: "Google Docs ドキュメントの本文を Markdown 形式またはプレーンテキストで取得する",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ readDocumentInput) (*mcp.CallToolResult, struct{}, error) {
		_ = docsService
		return nil, struct{}{}, errNotImplemented
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_documents",
		Description: "Google Drive 内の Google Docs ドキュメント一覧を取得する",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ listDocumentsInput) (*mcp.CallToolResult, struct{}, error) {
		_ = driveService
		return nil, struct{}{}, errNotImplemented
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_documents",
		Description: "キーワードで Google Docs ドキュメントを検索する",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ searchDocumentsInput) (*mcp.CallToolResult, struct{}, error) {
		_ = driveService
		return nil, struct{}{}, errNotImplemented
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_document_info",
		Description: "Google Docs ドキュメントのメタ情報（タイトル、作成日、更新日、オーナー等）を取得する",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ getDocumentInfoInput) (*mcp.CallToolResult, struct{}, error) {
		_ = driveService
		return nil, struct{}{}, errNotImplemented
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_comments",
		Description: "Google Docs ドキュメントのコメント一覧を取得する",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input listCommentsInput) (*mcp.CallToolResult, struct{}, error) {
		return listComments(ctx, driveService, input)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_comment",
		Description: "Google Docs ドキュメントの個別コメントと返信スレッドを取得する",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input getCommentInput) (*mcp.CallToolResult, struct{}, error) {
		return getComment(ctx, driveService, input)
	})
}
