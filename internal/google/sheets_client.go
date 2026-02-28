package google

import (
	"context"
	"net/http"

	"google.golang.org/api/option"
	sheets "google.golang.org/api/sheets/v4"
)

// SheetsClientImpl implements tools.SheetsClient using the real Google Sheets API.
type SheetsClientImpl struct {
	service *sheets.Service
}

// NewSheetsClient creates a new SheetsClientImpl from an authenticated HTTP client.
func NewSheetsClient(client *http.Client) (*SheetsClientImpl, error) {
	svc, err := sheets.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}
	return &SheetsClientImpl{service: svc}, nil
}

// GetSpreadsheet retrieves spreadsheet metadata by its ID.
func (s *SheetsClientImpl) GetSpreadsheet(ctx context.Context, spreadsheetID string) (*sheets.Spreadsheet, error) {
	return s.service.Spreadsheets.Get(spreadsheetID).Context(ctx).Do()
}

// GetValues retrieves cell values from a specified range using A1 notation.
func (s *SheetsClientImpl) GetValues(ctx context.Context, spreadsheetID, rangeA1 string) (*sheets.ValueRange, error) {
	return s.service.Spreadsheets.Values.Get(spreadsheetID, rangeA1).Context(ctx).Do()
}
