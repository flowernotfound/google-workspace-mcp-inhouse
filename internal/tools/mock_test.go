package tools

import (
	"context"

	docs "google.golang.org/api/docs/v1"
	drive "google.golang.org/api/drive/v3"
	sheets "google.golang.org/api/sheets/v4"
)

// mockDocsClient is a test double for DocsClient.
type mockDocsClient struct {
	getDocumentFn func(ctx context.Context, documentID string) (*docs.Document, error)
}

func (m *mockDocsClient) GetDocument(ctx context.Context, documentID string) (*docs.Document, error) {
	if m.getDocumentFn == nil {
		panic("mockDocsClient.GetDocument called but getDocumentFn is not set")
	}
	return m.getDocumentFn(ctx, documentID)
}

// mockDriveClient is a test double for DriveClient.
type mockDriveClient struct {
	listFilesFn    func(ctx context.Context, query, orderBy string, pageSize int64, fields string) (*drive.FileList, error)
	getFileFn      func(ctx context.Context, fileID, fields string) (*drive.File, error)
	listCommentsFn func(ctx context.Context, fileID, fields string, includeDeleted bool, pageSize int64, pageToken string) (*drive.CommentList, error)
	getCommentFn   func(ctx context.Context, fileID, commentID, fields string, includeDeleted bool) (*drive.Comment, error)
}

func (m *mockDriveClient) ListFiles(ctx context.Context, query, orderBy string, pageSize int64, fields string) (*drive.FileList, error) {
	if m.listFilesFn == nil {
		panic("mockDriveClient.ListFiles called but listFilesFn is not set")
	}
	return m.listFilesFn(ctx, query, orderBy, pageSize, fields)
}

func (m *mockDriveClient) GetFile(ctx context.Context, fileID, fields string) (*drive.File, error) {
	if m.getFileFn == nil {
		panic("mockDriveClient.GetFile called but getFileFn is not set")
	}
	return m.getFileFn(ctx, fileID, fields)
}

func (m *mockDriveClient) ListComments(ctx context.Context, fileID, fields string, includeDeleted bool, pageSize int64, pageToken string) (*drive.CommentList, error) {
	if m.listCommentsFn == nil {
		panic("mockDriveClient.ListComments called but listCommentsFn is not set")
	}
	return m.listCommentsFn(ctx, fileID, fields, includeDeleted, pageSize, pageToken)
}

func (m *mockDriveClient) GetComment(ctx context.Context, fileID, commentID, fields string, includeDeleted bool) (*drive.Comment, error) {
	if m.getCommentFn == nil {
		panic("mockDriveClient.GetComment called but getCommentFn is not set")
	}
	return m.getCommentFn(ctx, fileID, commentID, fields, includeDeleted)
}

// mockSheetsClient is a test double for SheetsClient.
type mockSheetsClient struct {
	getSpreadsheetFn func(ctx context.Context, spreadsheetID string) (*sheets.Spreadsheet, error)
	getValuesFn      func(ctx context.Context, spreadsheetID, rangeA1 string) (*sheets.ValueRange, error)
}

func (m *mockSheetsClient) GetSpreadsheet(ctx context.Context, spreadsheetID string) (*sheets.Spreadsheet, error) {
	if m.getSpreadsheetFn == nil {
		panic("mockSheetsClient.GetSpreadsheet called but getSpreadsheetFn is not set")
	}
	return m.getSpreadsheetFn(ctx, spreadsheetID)
}

func (m *mockSheetsClient) GetValues(ctx context.Context, spreadsheetID, rangeA1 string) (*sheets.ValueRange, error) {
	if m.getValuesFn == nil {
		panic("mockSheetsClient.GetValues called but getValuesFn is not set")
	}
	return m.getValuesFn(ctx, spreadsheetID, rangeA1)
}
