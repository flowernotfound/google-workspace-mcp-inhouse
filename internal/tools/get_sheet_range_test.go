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

func TestGetSheetRange_Success(t *testing.T) {
	mock := &mockSheetsClient{
		getValuesFn: func(_ context.Context, _ string, rangeA1 string) (*sheets.ValueRange, error) {
			assert.Equal(t, "Sheet1!A1:B2", rangeA1)
			return &sheets.ValueRange{
				Range: "Sheet1!A1:B2",
				Values: [][]interface{}{
					{"Name", "Age"},
					{"Alice", 30},
				},
			}, nil
		},
	}

	result := getSheetRange(context.Background(), mock, getSheetRangeInput{
		SpreadsheetID: "ss-id",
		Range:         "Sheet1!A1:B2",
	})
	assert.False(t, result.IsError)

	text := result.Content[0].(*mcp.TextContent).Text
	var res getSheetRangeResult
	require.NoError(t, json.Unmarshal([]byte(text), &res))
	assert.Equal(t, "ss-id", res.SpreadsheetID)
	assert.Equal(t, "Sheet1!A1:B2", res.Range)
	require.Len(t, res.Values, 2)
	assert.Len(t, res.Values[0], 2)
}

func TestGetSheetRange_EmptyValues(t *testing.T) {
	mock := &mockSheetsClient{
		getValuesFn: func(_ context.Context, _ string, _ string) (*sheets.ValueRange, error) {
			return &sheets.ValueRange{
				Range:  "Sheet1!A1:A1",
				Values: nil,
			}, nil
		},
	}

	result := getSheetRange(context.Background(), mock, getSheetRangeInput{
		SpreadsheetID: "ss-id",
		Range:         "Sheet1!A1:A1",
	})
	assert.False(t, result.IsError)

	text := result.Content[0].(*mcp.TextContent).Text
	var res getSheetRangeResult
	require.NoError(t, json.Unmarshal([]byte(text), &res))
	assert.Nil(t, res.Values)
}

func TestGetSheetRange_APIError(t *testing.T) {
	mock := &mockSheetsClient{
		getValuesFn: func(_ context.Context, _ string, _ string) (*sheets.ValueRange, error) {
			return nil, &googleapi.Error{Code: 404, Message: "not found"}
		},
	}

	result := getSheetRange(context.Background(), mock, getSheetRangeInput{
		SpreadsheetID: "ss-id",
		Range:         "Sheet1!A1:B2",
	})
	assert.True(t, result.IsError)
	assert.Contains(t, result.Content[0].(*mcp.TextContent).Text, "404")
}
