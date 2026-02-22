package google

import (
	"context"
	"net/http"

	docs "google.golang.org/api/docs/v1"
	"google.golang.org/api/option"
)

// NewDocsService creates a Google Docs API service from an authenticated HTTP client.
func NewDocsService(client *http.Client) (*docs.Service, error) {
	return docs.NewService(context.Background(), option.WithHTTPClient(client))
}
