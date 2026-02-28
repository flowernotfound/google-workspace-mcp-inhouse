package google

import (
	"context"
	"net/http"

	docs "google.golang.org/api/docs/v1"
	"google.golang.org/api/option"
)

// DocsClientImpl implements tools.DocsClient using the real Google Docs API.
type DocsClientImpl struct {
	service *docs.Service
}

// NewDocsClient creates a new DocsClientImpl from an authenticated HTTP client.
func NewDocsClient(client *http.Client) (*DocsClientImpl, error) {
	svc, err := docs.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}
	return &DocsClientImpl{service: svc}, nil
}

// GetDocument retrieves a Google Docs document by its ID.
func (d *DocsClientImpl) GetDocument(ctx context.Context, documentID string) (*docs.Document, error) {
	return d.service.Documents.Get(documentID).Context(ctx).Do()
}
