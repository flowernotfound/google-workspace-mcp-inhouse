package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Input types for each tool.

type readDocumentInput struct {
	DocumentID string  `json:"document_id" jsonschema:"required,the ID or URL of the Google Docs document"`
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
	DocumentID string `json:"document_id" jsonschema:"required,the ID or URL of the Google Docs document"`
}

type listCommentsInput struct {
	DocumentID      string `json:"document_id" jsonschema:"required,the ID or URL of the Google Docs document"`
	IncludeResolved bool   `json:"include_resolved" jsonschema:"whether to include resolved comments (default false)"`
}

type getCommentInput struct {
	DocumentID string `json:"document_id" jsonschema:"required,the ID or URL of the Google Docs document"`
	CommentID  string `json:"comment_id" jsonschema:"required,the comment ID"`
}

type readSpreadsheetInput struct {
	SpreadsheetID string  `json:"spreadsheet_id" jsonschema:"required,the ID or URL of the Google Sheets spreadsheet"`
	SheetName     *string `json:"sheet_name,omitempty" jsonschema:"sheet name to read (defaults to first sheet)"`
	Format        *string `json:"format,omitempty" jsonschema:"output format: 'csv' (default) or 'json'"`
}

type getSpreadsheetInfoInput struct {
	SpreadsheetID string `json:"spreadsheet_id" jsonschema:"required,the ID or URL of the Google Sheets spreadsheet"`
}

type listSpreadsheetsInput struct {
	FolderID   *string `json:"folder_id,omitempty" jsonschema:"folder ID to filter spreadsheets"`
	MaxResults *int    `json:"max_results,omitempty" jsonschema:"maximum number of results (default 20, max 100)"`
	OrderBy    *string `json:"order_by,omitempty" jsonschema:"sort order, e.g. 'modifiedTime desc' (default), 'name', 'createdTime'"`
}

type searchSpreadsheetsInput struct {
	Query      string `json:"query" jsonschema:"required,search keyword"`
	MaxResults *int   `json:"max_results,omitempty" jsonschema:"maximum number of results (default 10, max 50)"`
}

type getSheetRangeInput struct {
	SpreadsheetID string `json:"spreadsheet_id" jsonschema:"required,the ID or URL of the Google Sheets spreadsheet"`
	Range         string `json:"range" jsonschema:"required,range in A1 notation (e.g. Sheet1!A1:D10)"`
}

// RegisterTools registers all MCP tools on the server.
func RegisterTools(server *mcp.Server, docsClient DocsClient, driveClient DriveClient, sheetsClient SheetsClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "read_document",
		Description: "Google Docs ドキュメントの本文を Markdown 形式またはプレーンテキストで取得する",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input readDocumentInput) (*mcp.CallToolResult, any, error) {
		return readDocument(ctx, docsClient, input), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_documents",
		Description: "Google Drive 内の Google Docs ドキュメント一覧を取得する",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input listDocumentsInput) (*mcp.CallToolResult, any, error) {
		return listDocuments(ctx, driveClient, input), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_documents",
		Description: "キーワードで Google Docs ドキュメントを検索する",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input searchDocumentsInput) (*mcp.CallToolResult, any, error) {
		return searchDocuments(ctx, driveClient, input), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_document_info",
		Description: "Google Docs ドキュメントのメタ情報（タイトル、作成日、更新日、オーナー等）を取得する",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input getDocumentInfoInput) (*mcp.CallToolResult, any, error) {
		return getDocumentInfo(ctx, driveClient, input), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_comments",
		Description: "Google Docs ドキュメントのコメント一覧を取得する",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input listCommentsInput) (*mcp.CallToolResult, any, error) {
		return listComments(ctx, driveClient, input), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_comment",
		Description: "Google Docs ドキュメントの個別コメントと返信スレッドを取得する",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input getCommentInput) (*mcp.CallToolResult, any, error) {
		return getComment(ctx, driveClient, input), nil, nil
	})

	// Google Sheets tools

	mcp.AddTool(server, &mcp.Tool{
		Name:        "read_spreadsheet",
		Description: "Google Sheets スプレッドシートの内容を CSV または JSON 形式で取得する",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input readSpreadsheetInput) (*mcp.CallToolResult, any, error) {
		return readSpreadsheet(ctx, sheetsClient, input), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_spreadsheet_info",
		Description: "Google Sheets スプレッドシートのメタ情報（タイトル、シート一覧、ロケール、タイムゾーン等）を取得する",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input getSpreadsheetInfoInput) (*mcp.CallToolResult, any, error) {
		return getSpreadsheetInfo(ctx, sheetsClient, input), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_spreadsheets",
		Description: "Google Drive 内の Google Sheets スプレッドシート一覧を取得する",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input listSpreadsheetsInput) (*mcp.CallToolResult, any, error) {
		return listSpreadsheets(ctx, driveClient, input), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_spreadsheets",
		Description: "キーワードで Google Sheets スプレッドシートを検索する",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input searchSpreadsheetsInput) (*mcp.CallToolResult, any, error) {
		return searchSpreadsheets(ctx, driveClient, input), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_sheet_range",
		Description: "Google Sheets スプレッドシートの指定範囲（A1記法）のセル値を取得する",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input getSheetRangeInput) (*mcp.CallToolResult, any, error) {
		return getSheetRange(ctx, sheetsClient, input), nil, nil
	})
}
