package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/googleapi"
	sheets "google.golang.org/api/sheets/v4"
)

func minimalSpreadsheet(sheetNames ...string) *sheets.Spreadsheet {
	ss := make([]*sheets.Sheet, 0, len(sheetNames))
	for _, name := range sheetNames {
		ss = append(ss, &sheets.Sheet{
			Properties: &sheets.SheetProperties{Title: name},
		})
	}
	return &sheets.Spreadsheet{
		SpreadsheetId: "ss-id",
		Properties:    &sheets.SpreadsheetProperties{Title: "My Spreadsheet"},
		Sheets:        ss,
	}
}

func TestReadSpreadsheet_CSV(t *testing.T) {
	mock := &mockSheetsClient{
		getSpreadsheetFn: func(_ context.Context, _ string) (*sheets.Spreadsheet, error) {
			return minimalSpreadsheet("Sheet1"), nil
		},
		getValuesFn: func(_ context.Context, _ string, _ string) (*sheets.ValueRange, error) {
			return &sheets.ValueRange{
				Values: [][]interface{}{
					{"Name", "Age"},
					{"Alice", 30},
				},
			}, nil
		},
	}

	result := readSpreadsheet(context.Background(), mock, readSpreadsheetInput{
		SpreadsheetID: "ss-id",
	})
	assert.False(t, result.IsError)

	text := result.Content[0].(*mcp.TextContent).Text
	var res readSpreadsheetResult
	require.NoError(t, json.Unmarshal([]byte(text), &res))
	assert.Equal(t, "ss-id", res.SpreadsheetID)
	assert.Equal(t, "My Spreadsheet", res.Title)
	assert.Equal(t, "Sheet1", res.SheetName)
	assert.Equal(t, "csv", res.Format)
	assert.Equal(t, "Name,Age\nAlice,30", res.Content)
}

func TestReadSpreadsheet_JSON(t *testing.T) {
	mock := &mockSheetsClient{
		getSpreadsheetFn: func(_ context.Context, _ string) (*sheets.Spreadsheet, error) {
			return minimalSpreadsheet("Sheet1"), nil
		},
		getValuesFn: func(_ context.Context, _ string, _ string) (*sheets.ValueRange, error) {
			return &sheets.ValueRange{
				Values: [][]interface{}{
					{"Name", "Age"},
					{"Alice", 30},
				},
			}, nil
		},
	}

	format := "json"
	result := readSpreadsheet(context.Background(), mock, readSpreadsheetInput{
		SpreadsheetID: "ss-id",
		Format:        &format,
	})
	assert.False(t, result.IsError)

	text := result.Content[0].(*mcp.TextContent).Text
	var res readSpreadsheetResult
	require.NoError(t, json.Unmarshal([]byte(text), &res))
	assert.Equal(t, "json", res.Format)
	assert.Contains(t, res.Content, "Alice")
}

func TestReadSpreadsheet_DefaultsToFirstSheet(t *testing.T) {
	var capturedRange string
	mock := &mockSheetsClient{
		getSpreadsheetFn: func(_ context.Context, _ string) (*sheets.Spreadsheet, error) {
			return minimalSpreadsheet("First", "Second"), nil
		},
		getValuesFn: func(_ context.Context, _ string, rangeA1 string) (*sheets.ValueRange, error) {
			capturedRange = rangeA1
			return &sheets.ValueRange{Values: [][]interface{}{{"A"}}}, nil
		},
	}

	result := readSpreadsheet(context.Background(), mock, readSpreadsheetInput{
		SpreadsheetID: "ss-id",
	})
	assert.False(t, result.IsError)
	assert.Equal(t, "First", capturedRange)
}

func TestReadSpreadsheet_SpecifiedSheetName(t *testing.T) {
	var capturedRange string
	sheetName := "Custom"
	mock := &mockSheetsClient{
		getSpreadsheetFn: func(_ context.Context, _ string) (*sheets.Spreadsheet, error) {
			return minimalSpreadsheet("Default", "Custom"), nil
		},
		getValuesFn: func(_ context.Context, _ string, rangeA1 string) (*sheets.ValueRange, error) {
			capturedRange = rangeA1
			return &sheets.ValueRange{Values: [][]interface{}{{"A"}}}, nil
		},
	}

	result := readSpreadsheet(context.Background(), mock, readSpreadsheetInput{
		SpreadsheetID: "ss-id",
		SheetName:     &sheetName,
	})
	assert.False(t, result.IsError)
	assert.Equal(t, "Custom", capturedRange)
}

func TestReadSpreadsheet_NoSheets(t *testing.T) {
	mock := &mockSheetsClient{
		getSpreadsheetFn: func(_ context.Context, _ string) (*sheets.Spreadsheet, error) {
			return &sheets.Spreadsheet{
				SpreadsheetId: "ss-id",
				Properties:    &sheets.SpreadsheetProperties{Title: "Empty"},
				Sheets:        []*sheets.Sheet{},
			}, nil
		},
	}

	result := readSpreadsheet(context.Background(), mock, readSpreadsheetInput{
		SpreadsheetID: "ss-id",
	})
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "no sheets")
}

func TestReadSpreadsheet_UnsupportedFormat(t *testing.T) {
	mock := &mockSheetsClient{
		getSpreadsheetFn: func(_ context.Context, _ string) (*sheets.Spreadsheet, error) {
			return minimalSpreadsheet("Sheet1"), nil
		},
		getValuesFn: func(_ context.Context, _ string, _ string) (*sheets.ValueRange, error) {
			return &sheets.ValueRange{Values: [][]interface{}{{"A"}}}, nil
		},
	}

	format := "xml"
	result := readSpreadsheet(context.Background(), mock, readSpreadsheetInput{
		SpreadsheetID: "ss-id",
		Format:        &format,
	})
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "unsupported format")
}

func TestReadSpreadsheet_APIError(t *testing.T) {
	mock := &mockSheetsClient{
		getSpreadsheetFn: func(_ context.Context, _ string) (*sheets.Spreadsheet, error) {
			return nil, &googleapi.Error{Code: 404, Message: "not found"}
		},
	}

	result := readSpreadsheet(context.Background(), mock, readSpreadsheetInput{
		SpreadsheetID: "ss-id",
	})
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "404")
}
