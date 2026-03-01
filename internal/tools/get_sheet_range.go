package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type getSheetRangeResult struct {
	SpreadsheetID string          `json:"spreadsheet_id"`
	Range         string          `json:"range"`
	Values        [][]interface{} `json:"values"`
}

func getSheetRange(ctx context.Context, sheetsClient SheetsClient, input getSheetRangeInput) *mcp.CallToolResult {
	spreadsheetID := ResolveID(strings.TrimSpace(input.SpreadsheetID))
	rng := strings.TrimSpace(input.Range)

	if spreadsheetID == "" {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{Text: "spreadsheet_id must not be empty"}},
		}
	}
	if rng == "" {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{Text: "range must not be empty"}},
		}
	}

	values, err := sheetsClient.GetValues(ctx, spreadsheetID, rng)
	if err != nil {
		return errorResult(err)
	}

	result := getSheetRangeResult{
		SpreadsheetID: spreadsheetID,
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
