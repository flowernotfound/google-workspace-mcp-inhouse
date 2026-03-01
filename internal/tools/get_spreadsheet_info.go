package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type spreadsheetInfo struct {
	SpreadsheetID string      `json:"spreadsheet_id"`
	Title         string      `json:"title"`
	Locale        string      `json:"locale"`
	TimeZone      string      `json:"timezone"`
	Sheets        []sheetInfo `json:"sheets"`
}

type sheetInfo struct {
	SheetID  int64  `json:"sheet_id"`
	Title    string `json:"title"`
	Index    int64  `json:"index"`
	RowCount int64  `json:"row_count"`
	ColCount int64  `json:"col_count"`
}

func getSpreadsheetInfo(ctx context.Context, sheetsClient SheetsClient, input getSpreadsheetInfoInput) *mcp.CallToolResult {
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

	sheetsList := make([]sheetInfo, 0, len(spreadsheet.Sheets))
	for _, s := range spreadsheet.Sheets {
		if s.Properties == nil {
			continue
		}
		si := sheetInfo{
			SheetID: s.Properties.SheetId,
			Title:   s.Properties.Title,
			Index:   s.Properties.Index,
		}
		if s.Properties.GridProperties != nil {
			si.RowCount = s.Properties.GridProperties.RowCount
			si.ColCount = s.Properties.GridProperties.ColumnCount
		}
		sheetsList = append(sheetsList, si)
	}

	info := spreadsheetInfo{
		SpreadsheetID: spreadsheet.SpreadsheetId,
		Title:         spreadsheet.Properties.Title,
		Locale:        spreadsheet.Properties.Locale,
		TimeZone:      spreadsheet.Properties.TimeZone,
		Sheets:        sheetsList,
	}

	data, err := json.Marshal(info)
	if err != nil {
		return errorResult(fmt.Errorf("failed to serialize response: %w", err))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}
}
