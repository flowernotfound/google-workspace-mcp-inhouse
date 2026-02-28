package converter

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FormatValuesAsCSV converts a 2D array of cell values to CSV format.
// Fields containing commas, double quotes, or newlines are properly escaped per RFC 4180.
func FormatValuesAsCSV(values [][]interface{}) string {
	lines := make([]string, 0, len(values))
	for _, row := range values {
		cells := make([]string, 0, len(row))
		for _, cell := range row {
			str := fmt.Sprintf("%v", cell)
			if strings.ContainsAny(str, ",\"\n\r") {
				str = "\"" + strings.ReplaceAll(str, "\"", "\"\"") + "\""
			}
			cells = append(cells, str)
		}
		lines = append(lines, strings.Join(cells, ","))
	}
	return strings.Join(lines, "\n")
}

// FormatValuesAsJSON converts a 2D array to a JSON array of objects.
// The first row is used as headers (keys), remaining rows become objects.
func FormatValuesAsJSON(values [][]interface{}) (string, error) {
	if len(values) == 0 {
		return "[]", nil
	}

	headers := make([]string, len(values[0]))
	for i, h := range values[0] {
		headers[i] = fmt.Sprintf("%v", h)
	}

	if len(values) == 1 {
		return "[]", nil
	}

	result := make([]map[string]interface{}, 0, len(values)-1)
	for _, row := range values[1:] {
		obj := make(map[string]interface{})
		for i, cell := range row {
			if i < len(headers) {
				obj[headers[i]] = cell
			}
		}
		result = append(result, obj)
	}

	data, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
