package tools

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestServerAndClient sets up an in-memory MCP server with all tools registered
// and returns a connected client session for use in tests.
func newTestServerAndClient(t *testing.T) *mcp.ClientSession {
	t.Helper()

	ctx := context.Background()

	server := mcp.NewServer(&mcp.Implementation{Name: "test-server", Version: "0.0.0"}, nil)
	RegisterTools(server, nil, nil, nil)

	serverTransport, clientTransport := mcp.NewInMemoryTransports()

	_, err := server.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.0"}, nil)
	cs, err := client.Connect(ctx, clientTransport, nil)
	require.NoError(t, err)

	t.Cleanup(func() { cs.Close() })

	return cs
}

func TestRegisterTools_ToolCount(t *testing.T) {
	cs := newTestServerAndClient(t)

	res, err := cs.ListTools(context.Background(), nil)
	require.NoError(t, err)

	assert.Len(t, res.Tools, 11)
}

func TestRegisterTools_ToolNames(t *testing.T) {
	cs := newTestServerAndClient(t)

	res, err := cs.ListTools(context.Background(), nil)
	require.NoError(t, err)

	names := make([]string, len(res.Tools))
	for i, tool := range res.Tools {
		names[i] = tool.Name
	}

	expected := []string{
		"read_document",
		"list_documents",
		"search_documents",
		"get_document_info",
		"list_comments",
		"get_comment",
		"read_spreadsheet",
		"get_spreadsheet_info",
		"list_spreadsheets",
		"search_spreadsheets",
		"get_sheet_range",
	}
	assert.ElementsMatch(t, expected, names)
}
