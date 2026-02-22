package main

import (
	"context"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/flowernotfound/google-workspace-mcp-inhouse/internal/auth"
	internalgoogle "github.com/flowernotfound/google-workspace-mcp-inhouse/internal/google"
	"github.com/flowernotfound/google-workspace-mcp-inhouse/internal/tools"
)

func main() {
	log.SetOutput(os.Stderr)

	if len(os.Args) > 1 && os.Args[1] == "auth" {
		if err := auth.RunAuthFlow(context.Background()); err != nil {
			log.Fatalf("auth error: %v", err)
		}
		return
	}

	// MCP server mode
	client, err := auth.Authorize(context.Background())
	if err != nil {
		log.Fatalf("authentication error: %v\nRun `google-workspace-mcp-inhouse auth` to authenticate", err)
	}

	docsService, err := internalgoogle.NewDocsService(client)
	if err != nil {
		log.Fatalf("failed to initialize Docs API client: %v", err)
	}

	driveService, err := internalgoogle.NewDriveService(client)
	if err != nil {
		log.Fatalf("failed to initialize Drive API client: %v", err)
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "google-workspace-mcp-inhouse",
		Version: "1.0.0",
	}, nil)

	tools.RegisterTools(server, docsService, driveService)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("MCP server error: %v", err)
	}
}
