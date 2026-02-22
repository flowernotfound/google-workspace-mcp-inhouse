package google

import (
	"context"
	"net/http"

	drive "google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// NewDriveService creates a Google Drive API service from an authenticated HTTP client.
func NewDriveService(client *http.Client) (*drive.Service, error) {
	return drive.NewService(context.Background(), option.WithHTTPClient(client))
}
