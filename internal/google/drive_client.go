package google

import (
	"context"
	"net/http"

	drive "google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

// DriveClientImpl implements tools.DriveClient using the real Google Drive API.
type DriveClientImpl struct {
	service *drive.Service
}

// NewDriveClient creates a new DriveClientImpl from an authenticated HTTP client.
func NewDriveClient(client *http.Client) (*DriveClientImpl, error) {
	svc, err := drive.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}
	return &DriveClientImpl{service: svc}, nil
}

// ListFiles lists files matching the given query.
func (d *DriveClientImpl) ListFiles(ctx context.Context, query, orderBy string, pageSize int64, fields string) (*drive.FileList, error) {
	req := d.service.Files.List().
		Q(query).
		PageSize(pageSize).
		Fields(googleapi.Field(fields)).
		Context(ctx)
	if orderBy != "" {
		req = req.OrderBy(orderBy)
	}
	return req.Do()
}

// GetFile retrieves file metadata by its ID.
func (d *DriveClientImpl) GetFile(ctx context.Context, fileID, fields string) (*drive.File, error) {
	return d.service.Files.Get(fileID).
		Fields(googleapi.Field(fields)).
		Context(ctx).
		Do()
}

// ListComments lists comments on a file, supporting pagination.
func (d *DriveClientImpl) ListComments(ctx context.Context, fileID, fields string, includeDeleted bool, pageSize int64, pageToken string) (*drive.CommentList, error) {
	req := d.service.Comments.List(fileID).
		Fields(googleapi.Field(fields)).
		IncludeDeleted(includeDeleted).
		PageSize(pageSize).
		Context(ctx)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Do()
}

// GetComment retrieves a single comment by file ID and comment ID.
func (d *DriveClientImpl) GetComment(ctx context.Context, fileID, commentID, fields string, includeDeleted bool) (*drive.Comment, error) {
	return d.service.Comments.Get(fileID, commentID).
		Fields(googleapi.Field(fields)).
		IncludeDeleted(includeDeleted).
		Context(ctx).
		Do()
}
