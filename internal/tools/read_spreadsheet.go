package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/flowernotfound/google-workspace-mcp-inhouse/internal/converter"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// quoteSheetName wraps a sheet name in single quotes for A1 notation.
// Single quotes within the name are escaped by doubling them.
// This ensures sheet names containing spaces, '!', or other special characters
// are correctly interpreted by the Google Sheets API.
func quoteSheetName(name string) string {
	escaped := strings.ReplaceAll(name, "'", "''")
	return "'" + escaped + "'"
}

type readSpreadsheetResult struct {
	SpreadsheetID string `json:"spreadsheet_id"`
	Title         string `json:"title"`
	SheetName     string `json:"sheet_name"`
	Format        string `json:"format"`
	Content       string `json:"content"`
}

func readSpreadsheet(ctx context.Context, sheetsClient SheetsClient, input readSpreadsheetInput) *mcp.CallToolResult {
	// Validate format before making any API calls.
	format := "csv"
	if input.Format != nil && *input.Format != "" {
		format = *input.Format
	}
	switch format {
	case "csv", "json":
		// valid
	default:
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("unsupported format %q: use 'csv' or 'json'", format)}},
		}
	}

	spreadsheet, err := sheetsClient.GetSpreadsheet(ctx, input.SpreadsheetID)
	if err != nil {
		return errorResult(err)
	}

	if spreadsheet.Properties == nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{Text: "spreadsheet metadata is missing"}},
		}
	}

	var sheetName string
	if input.SheetName != nil && *input.SheetName != "" {
		sheetName = *input.SheetName
	} else if len(spreadsheet.Sheets) > 0 {
		if spreadsheet.Sheets[0].Properties == nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: "first sheet has no properties"}},
			}
		}
		sheetName = spreadsheet.Sheets[0].Properties.Title
	} else {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{Text: "spreadsheet has no sheets"}},
		}
	}

	values, err := sheetsClient.GetValues(ctx, input.SpreadsheetID, quoteSheetName(sheetName))
	if err != nil {
		return errorResult(err)
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
		// unreachable: format is validated at the top of this function
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("unsupported format %q", format)}},
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
