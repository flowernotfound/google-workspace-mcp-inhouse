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

func TestGetSpreadsheetInfo_Success(t *testing.T) {
	mock := &mockSheetsClient{
		getSpreadsheetFn: func(_ context.Context, _ string) (*sheets.Spreadsheet, error) {
			return &sheets.Spreadsheet{
				SpreadsheetId: "ss-id",
				Properties: &sheets.SpreadsheetProperties{
					Title:    "Test Spreadsheet",
					Locale:   "ja_JP",
					TimeZone: "Asia/Tokyo",
				},
				Sheets: []*sheets.Sheet{
					{
						Properties: &sheets.SheetProperties{
							SheetId: 0,
							Title:   "Sheet1",
							Index:   0,
							GridProperties: &sheets.GridProperties{
								RowCount:    100,
								ColumnCount: 26,
							},
						},
					},
					{
						Properties: &sheets.SheetProperties{
							SheetId: 123,
							Title:   "Sheet2",
							Index:   1,
							GridProperties: &sheets.GridProperties{
								RowCount:    50,
								ColumnCount: 10,
							},
						},
					},
				},
			}, nil
		},
	}

	result := getSpreadsheetInfo(context.Background(), mock, getSpreadsheetInfoInput{
		SpreadsheetID: "ss-id",
	})
	assert.False(t, result.IsError)

	text := result.Content[0].(*mcp.TextContent).Text
	var info spreadsheetInfo
	require.NoError(t, json.Unmarshal([]byte(text), &info))
	assert.Equal(t, "ss-id", info.SpreadsheetID)
	assert.Equal(t, "Test Spreadsheet", info.Title)
	assert.Equal(t, "ja_JP", info.Locale)
	assert.Equal(t, "Asia/Tokyo", info.TimeZone)
	require.Len(t, info.Sheets, 2)
	assert.Equal(t, "Sheet1", info.Sheets[0].Title)
	assert.Equal(t, int64(100), info.Sheets[0].RowCount)
	assert.Equal(t, int64(26), info.Sheets[0].ColCount)
	assert.Equal(t, "Sheet2", info.Sheets[1].Title)
}

func TestGetSpreadsheetInfo_NoGridProperties(t *testing.T) {
	mock := &mockSheetsClient{
		getSpreadsheetFn: func(_ context.Context, _ string) (*sheets.Spreadsheet, error) {
			return &sheets.Spreadsheet{
				SpreadsheetId: "ss-id",
				Properties:    &sheets.SpreadsheetProperties{Title: "Test"},
				Sheets: []*sheets.Sheet{
					{
						Properties: &sheets.SheetProperties{
							Title: "Sheet1",
							// GridProperties is nil
						},
					},
				},
			}, nil
		},
	}

	result := getSpreadsheetInfo(context.Background(), mock, getSpreadsheetInfoInput{
		SpreadsheetID: "ss-id",
	})
	assert.False(t, result.IsError)

	text := result.Content[0].(*mcp.TextContent).Text
	var info spreadsheetInfo
	require.NoError(t, json.Unmarshal([]byte(text), &info))
	assert.Equal(t, int64(0), info.Sheets[0].RowCount)
	assert.Equal(t, int64(0), info.Sheets[0].ColCount)
}

func TestGetSpreadsheetInfo_NilProperties(t *testing.T) {
	mock := &mockSheetsClient{
		getSpreadsheetFn: func(_ context.Context, _ string) (*sheets.Spreadsheet, error) {
			return &sheets.Spreadsheet{
				SpreadsheetId: "ss-id",
				Properties:    nil,
			}, nil
		},
	}

	result := getSpreadsheetInfo(context.Background(), mock, getSpreadsheetInfoInput{
		SpreadsheetID: "ss-id",
	})
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "metadata is missing")
}

func TestGetSpreadsheetInfo_NilSheetProperties(t *testing.T) {
	mock := &mockSheetsClient{
		getSpreadsheetFn: func(_ context.Context, _ string) (*sheets.Spreadsheet, error) {
			return &sheets.Spreadsheet{
				SpreadsheetId: "ss-id",
				Properties:    &sheets.SpreadsheetProperties{Title: "Test"},
				Sheets: []*sheets.Sheet{
					{Properties: nil},
					{Properties: &sheets.SheetProperties{Title: "Valid"}},
				},
			}, nil
		},
	}

	result := getSpreadsheetInfo(context.Background(), mock, getSpreadsheetInfoInput{
		SpreadsheetID: "ss-id",
	})
	assert.False(t, result.IsError)

	text := result.Content[0].(*mcp.TextContent).Text
	var info spreadsheetInfo
	require.NoError(t, json.Unmarshal([]byte(text), &info))
	// Nil-properties sheet is skipped, only "Valid" remains
	require.Len(t, info.Sheets, 1)
	assert.Equal(t, "Valid", info.Sheets[0].Title)
}

func TestGetSpreadsheetInfo_APIError(t *testing.T) {
	mock := &mockSheetsClient{
		getSpreadsheetFn: func(_ context.Context, _ string) (*sheets.Spreadsheet, error) {
			return nil, &googleapi.Error{Code: 403, Message: "access denied"}
		},
	}

	result := getSpreadsheetInfo(context.Background(), mock, getSpreadsheetInfoInput{
		SpreadsheetID: "ss-id",
	})
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "403")
}
