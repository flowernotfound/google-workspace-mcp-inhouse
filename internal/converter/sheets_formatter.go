package converter

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
)

// FormatValuesAsCSV converts a 2D array of cell values to CSV format.
// Rows are padded to the maximum column count to produce a rectangular CSV.
// Fields containing commas, double quotes, or newlines are escaped per RFC 4180.
// Nil cells are output as empty strings.
func FormatValuesAsCSV(values [][]interface{}) string {
	if len(values) == 0 {
		return ""
	}

	maxCols := 0
	for _, row := range values {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}

	var buf strings.Builder
	w := csv.NewWriter(&buf)
	for _, row := range values {
		record := make([]string, maxCols)
		for i := 0; i < maxCols; i++ {
			if i < len(row) && row[i] != nil {
				record[i] = fmt.Sprintf("%v", row[i])
			}
		}
		w.Write(record) //nolint:errcheck // writing to strings.Builder never fails
	}
	w.Flush()
	return strings.TrimRight(buf.String(), "\n")
}

// FormatValuesAsJSON converts a 2D array to a JSON array of objects.
// The first row is used as headers (keys), remaining rows become objects.
func FormatValuesAsJSON(values [][]interface{}) (string, error) {
	if len(values) == 0 {
		return "[]", nil
	}

	headers := make([]string, len(values[0]))
	seen := make(map[string]int, len(values[0]))
	for i, h := range values[0] {
		var name string
		if h == nil || fmt.Sprintf("%v", h) == "" {
			name = fmt.Sprintf("column_%d", i+1)
		} else {
			name = fmt.Sprintf("%v", h)
		}
		if count, exists := seen[name]; exists {
			seen[name] = count + 1
			name = fmt.Sprintf("%s_%d", name, count+1)
		} else {
			seen[name] = 1
		}
		headers[i] = name
	}

	if len(values) == 1 {
		return "[]", nil
	}

	result := make([]map[string]interface{}, 0, len(values)-1)
	for _, row := range values[1:] {
		obj := make(map[string]interface{}, len(headers))
		for _, h := range headers {
			obj[h] = nil
		}
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
