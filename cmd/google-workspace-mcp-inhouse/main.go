package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/flowernotfound/google-workspace-mcp-inhouse/internal/auth"
	internalgoogle "github.com/flowernotfound/google-workspace-mcp-inhouse/internal/google"
	"github.com/flowernotfound/google-workspace-mcp-inhouse/internal/tools"
	"github.com/flowernotfound/google-workspace-mcp-inhouse/internal/updater"
)

// version is set by GoReleaser via ldflags at build time (e.g. "v0.1.42").
var version = "dev"

func main() {
	log.SetOutput(os.Stderr)

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "auth":
			if err := auth.RunAuthFlow(context.Background()); err != nil {
				log.Fatalf("auth error: %v", err)
			}
			return
		case "update":
			if err := updater.Run(context.Background(), version); err != nil {
				log.Fatalf("update error: %v", err)
			}
			return
		case "--version", "version":
			fmt.Println(version)
			return
		}
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
		Version: version,
	}, &mcp.ServerOptions{
		InitializedHandler: func(ctx context.Context, req *mcp.InitializedRequest) {
			go checkForUpdate(ctx, req.Session)
		},
	})

	tools.RegisterTools(server, docsService, driveService)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("MCP server error: %v", err)
	}
}

const updateCheckTimeout = 5 * time.Second

// checkForUpdate queries GitHub for newer releases and notifies via stderr and
// MCP logging notification. Errors are logged but never block the server.
func checkForUpdate(ctx context.Context, session *mcp.ServerSession) {
	checkCtx, cancel := context.WithTimeout(context.Background(), updateCheckTimeout)
	defer cancel()

	latest, hasUpdate, err := updater.CheckUpdate(checkCtx, version)
	if err != nil {
		log.Printf("[updater] update check failed: %v", err)
		return
	}
	if !hasUpdate {
		return
	}

	msg := fmt.Sprintf("New version %s is available (current: %s). Run: google-workspace-mcp-inhouse update", latest, version)

	// stderr fallback (always written to client log files)
	log.Printf("[updater] %s", msg)

	// MCP notifications/message (delivered if client has called logging/setLevel)
	_ = session.Log(ctx, &mcp.LoggingMessageParams{
		Level:  "warning",
		Logger: "updater",
		Data:   msg,
	})
}
