package tools

import (
	"context"

	docs "google.golang.org/api/docs/v1"
	drive "google.golang.org/api/drive/v3"
)

// DocsClient abstracts Google Docs API operations used by MCP tools.
type DocsClient interface {
	// GetDocument retrieves a Google Docs document by its ID.
	GetDocument(ctx context.Context, documentID string) (*docs.Document, error)
}

// DriveClient abstracts Google Drive API operations used by MCP tools.
type DriveClient interface {
	// ListFiles lists files matching the given query.
	ListFiles(ctx context.Context, query, orderBy string, pageSize int64, fields string) (*drive.FileList, error)

	// GetFile retrieves file metadata by its ID.
	GetFile(ctx context.Context, fileID, fields string) (*drive.File, error)

	// ListComments lists comments on a file, supporting pagination.
	ListComments(ctx context.Context, fileID, fields string, includeDeleted bool, pageSize int64, pageToken string) (*drive.CommentList, error)

	// GetComment retrieves a single comment by file ID and comment ID.
	GetComment(ctx context.Context, fileID, commentID, fields string, includeDeleted bool) (*drive.Comment, error)
}
