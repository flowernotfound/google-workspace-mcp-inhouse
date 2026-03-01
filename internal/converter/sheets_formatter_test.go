package converter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatValuesAsCSV(t *testing.T) {
	tests := []struct {
		name     string
		values   [][]interface{}
		expected string
	}{
		{
			name: "basic values",
			values: [][]interface{}{
				{"Name", "Age", "City"},
				{"Alice", 30, "Tokyo"},
				{"Bob", 25, "Osaka"},
			},
			expected: "Name,Age,City\nAlice,30,Tokyo\nBob,25,Osaka",
		},
		{
			name: "field containing comma",
			values: [][]interface{}{
				{"Name", "Address"},
				{"Alice", "Tokyo, Japan"},
			},
			expected: "Name,Address\nAlice,\"Tokyo, Japan\"",
		},
		{
			name: "field containing double quote",
			values: [][]interface{}{
				{"Name", "Note"},
				{"Alice", `She said "hello"`},
			},
			expected: "Name,Note\nAlice,\"She said \"\"hello\"\"\"",
		},
		{
			name: "field containing newline",
			values: [][]interface{}{
				{"Name", "Bio"},
				{"Alice", "Line1\nLine2"},
			},
			expected: "Name,Bio\nAlice,\"Line1\nLine2\"",
		},
		{
			name:     "empty values",
			values:   [][]interface{}{},
			expected: "",
		},
		{
			name:     "nil values",
			values:   nil,
			expected: "",
		},
		{
			name: "single row",
			values: [][]interface{}{
				{"A", "B", "C"},
			},
			expected: "A,B,C",
		},
		{
			name: "nil cells output as empty strings",
			values: [][]interface{}{
				{"Name", "Age"},
				{"Alice", nil},
				{nil, 30},
			},
			expected: "Name,Age\nAlice,\n,30",
		},
		{
			name: "non-rectangular data padded to max columns",
			values: [][]interface{}{
				{"Name", "Age", "City", "Country"},
				{"Alice", 30, "Tokyo", "Japan"},
				{"Bob", 25},
				{"Charlie"},
			},
			expected: "Name,Age,City,Country\nAlice,30,Tokyo,Japan\nBob,25,,\nCharlie,,,",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatValuesAsCSV(tt.values)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatValuesAsJSON(t *testing.T) {
	tests := []struct {
		name     string
		values   [][]interface{}
		expected string
		wantErr  bool
	}{
		{
			name:     "empty values",
			values:   [][]interface{}{},
			expected: "[]",
		},
		{
			name:     "nil values",
			values:   nil,
			expected: "[]",
		},
		{
			name: "headers only",
			values: [][]interface{}{
				{"Name", "Age"},
			},
			expected: "[]",
		},
		{
			name: "basic conversion",
			values: [][]interface{}{
				{"Name", "Age"},
				{"Alice", 30},
				{"Bob", 25},
			},
		},
		{
			name: "row with fewer columns than headers",
			values: [][]interface{}{
				{"A", "B", "C"},
				{"1"},
			},
		},
		{
			name: "row with more columns than headers",
			values: [][]interface{}{
				{"A", "B"},
				{"1", "2", "3"},
			},
		},
		{
			name: "nil cells in data rows",
			values: [][]interface{}{
				{"Name", "Age"},
				{"Alice", nil},
			},
		},
		{
			name: "nil header cell",
			values: [][]interface{}{
				{"Name", nil},
				{"Alice", 30},
			},
		},
		{
			name: "duplicate headers get suffix",
			values: [][]interface{}{
				{"Name", "Score", "Score"},
				{"Alice", 90, 85},
			},
		},
		{
			name: "multiple nil headers get column_N names",
			values: [][]interface{}{
				{nil, nil, "Name"},
				{"X", "Y", "Alice"},
			},
		},
		{
			name: "triple duplicate headers",
			values: [][]interface{}{
				{"Val", "Val", "Val"},
				{"a", "b", "c"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatValuesAsJSON(tt.values)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			if tt.expected != "" {
				assert.Equal(t, tt.expected, result)
				return
			}

			// For complex cases, parse JSON and verify structure
			var parsed []map[string]interface{}
			require.NoError(t, json.Unmarshal([]byte(result), &parsed))

			switch tt.name {
			case "basic conversion":
				require.Len(t, parsed, 2)
				assert.Equal(t, "Alice", parsed[0]["Name"])
				assert.Equal(t, float64(30), parsed[0]["Age"])
				assert.Equal(t, "Bob", parsed[1]["Name"])
				assert.Equal(t, float64(25), parsed[1]["Age"])
			case "row with fewer columns than headers":
				require.Len(t, parsed, 1)
				assert.Equal(t, "1", parsed[0]["A"])
				assert.Nil(t, parsed[0]["B"])
				assert.Nil(t, parsed[0]["C"])
				// All header keys must be present
				assert.Len(t, parsed[0], 3)
			case "row with more columns than headers":
				require.Len(t, parsed, 1)
				assert.Equal(t, "1", parsed[0]["A"])
				assert.Equal(t, "2", parsed[0]["B"])
				// Extra column "3" should be ignored
				assert.Len(t, parsed[0], 2)
			case "nil cells in data rows":
				require.Len(t, parsed, 1)
				assert.Equal(t, "Alice", parsed[0]["Name"])
				assert.Nil(t, parsed[0]["Age"])
				// Both keys must be present
				assert.Len(t, parsed[0], 2)
			case "nil header cell":
				require.Len(t, parsed, 1)
				assert.Equal(t, "Alice", parsed[0]["Name"])
				// nil header becomes "column_2"
				assert.Equal(t, float64(30), parsed[0]["column_2"])
			case "duplicate headers get suffix":
				require.Len(t, parsed, 1)
				assert.Equal(t, "Alice", parsed[0]["Name"])
				assert.Equal(t, float64(90), parsed[0]["Score"])
				assert.Equal(t, float64(85), parsed[0]["Score_2"])
				assert.Len(t, parsed[0], 3)
			case "multiple nil headers get column_N names":
				require.Len(t, parsed, 1)
				assert.Equal(t, "X", parsed[0]["column_1"])
				assert.Equal(t, "Y", parsed[0]["column_2"])
				assert.Equal(t, "Alice", parsed[0]["Name"])
			case "triple duplicate headers":
				require.Len(t, parsed, 1)
				assert.Equal(t, "a", parsed[0]["Val"])
				assert.Equal(t, "b", parsed[0]["Val_2"])
				assert.Equal(t, "c", parsed[0]["Val_3"])
			}
		})
	}
}
