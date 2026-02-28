package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type getSheetRangeResult struct {
	SpreadsheetID string          `json:"spreadsheet_id"`
	Range         string          `json:"range"`
	Values        [][]interface{} `json:"values"`
}

func getSheetRange(ctx context.Context, sheetsClient SheetsClient, input getSheetRangeInput) *mcp.CallToolResult {
	values, err := sheetsClient.GetValues(ctx, input.SpreadsheetID, input.Range)
	if err != nil {
		return errorResult(err)
	}

	result := getSheetRangeResult{
		SpreadsheetID: input.SpreadsheetID,
		Range:         values.Range,
		Values:        values.Values,
	}

	data, err := json.Marshal(result)
	if err != nil {
		return errorResult(fmt.Errorf("failed to serialize response: %w", err))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}
}
