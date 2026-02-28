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
			case "row with more columns than headers":
				require.Len(t, parsed, 1)
				assert.Equal(t, "1", parsed[0]["A"])
				assert.Equal(t, "2", parsed[0]["B"])
				// Extra column "3" should be ignored
				assert.Len(t, parsed[0], 2)
			}
		})
	}
}
