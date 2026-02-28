package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/flowernotfound/google-workspace-mcp-inhouse/internal/converter"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type readSpreadsheetResult struct {
	SpreadsheetID string `json:"spreadsheet_id"`
	Title         string `json:"title"`
	SheetName     string `json:"sheet_name"`
	Format        string `json:"format"`
	Content       string `json:"content"`
}

func readSpreadsheet(ctx context.Context, sheetsClient SheetsClient, input readSpreadsheetInput) *mcp.CallToolResult {
	spreadsheet, err := sheetsClient.GetSpreadsheet(ctx, input.SpreadsheetID)
	if err != nil {
		return errorResult(err)
	}

	var sheetName string
	if input.SheetName != nil && *input.SheetName != "" {
		sheetName = *input.SheetName
	} else if len(spreadsheet.Sheets) > 0 {
		sheetName = spreadsheet.Sheets[0].Properties.Title
	} else {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{Text: "spreadsheet has no sheets"}},
		}
	}

	values, err := sheetsClient.GetValues(ctx, input.SpreadsheetID, sheetName)
	if err != nil {
		return errorResult(err)
	}

	format := "csv"
	if input.Format != nil && *input.Format != "" {
		format = *input.Format
	}

	var content string
	switch format {
	case "csv":
		content = converter.FormatValuesAsCSV(values.Values)
	case "json":
		content, err = converter.FormatValuesAsJSON(values.Values)
		if err != nil {
			return errorResult(fmt.Errorf("failed to format as JSON: %w", err))
		}
	default:
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("unsupported format %q: use 'csv' or 'json'", format)}},
		}
	}

	result := readSpreadsheetResult{
		SpreadsheetID: spreadsheet.SpreadsheetId,
		Title:         spreadsheet.Properties.Title,
		SheetName:     sheetName,
		Format:        format,
		Content:       content,
	}
	data, err := json.Marshal(result)
	if err != nil {
		return errorResult(fmt.Errorf("failed to serialize response: %w", err))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}
}
