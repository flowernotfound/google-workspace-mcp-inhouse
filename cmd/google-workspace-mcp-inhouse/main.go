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
		case "help", "--help", "-h":
			printUsage()
			return
		default:
			msg := fmt.Sprintf("unknown command: %q\n\n", os.Args[1])
			fmt.Fprint(os.Stderr, msg)
			printUsage()
			os.Exit(1)
		}
	}

	// MCP server mode
	client, err := auth.Authorize(context.Background())
	if err != nil {
		log.Fatalf("authentication error: %v\nRun `google-workspace-mcp-inhouse auth` to authenticate", err)
	}

	docsClient, err := internalgoogle.NewDocsClient(client)
	if err != nil {
		log.Fatalf("failed to initialize Docs API client: %v", err)
	}

	driveClient, err := internalgoogle.NewDriveClient(client)
	if err != nil {
		log.Fatalf("failed to initialize Drive API client: %v", err)
	}

	sheetsClient, err := internalgoogle.NewSheetsClient(client)
	if err != nil {
		log.Fatalf("failed to initialize Sheets API client: %v", err)
	}

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "google-workspace-mcp-inhouse",
		Version: version,
	}, &mcp.ServerOptions{
		InitializedHandler: func(_ context.Context, req *mcp.InitializedRequest) {
			go checkForUpdate(req.Session) //nolint:gosec // G118: intentionally uses context.Background inside goroutine because InitializedHandler ctx may be cancelled
		},
	})

	tools.RegisterTools(server, docsClient, driveClient, sheetsClient)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("MCP server error: %v", err)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  google-workspace-mcp-inhouse             Start the MCP server (stdio mode)\n")
	fmt.Fprintf(os.Stderr, "  google-workspace-mcp-inhouse [command]   Run a subcommand\n\n")
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  auth                  Run OAuth 2.0 authentication flow\n")
	fmt.Fprintf(os.Stderr, "  update                Check for and install updates\n")
	fmt.Fprintf(os.Stderr, "  version, --version    Show version\n")
	fmt.Fprintf(os.Stderr, "  help, --help, -h      Show this help message\n")
}

const updateCheckTimeout = 5 * time.Second

// checkForUpdate queries GitHub for newer releases and notifies via stderr and
// MCP logging notification. Errors are logged but never block the server.
func checkForUpdate(session *mcp.ServerSession) {
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

	// MCP notifications/message (delivered if client has called logging/setLevel).
	// Use a fresh context because the InitializedHandler's ctx may be cancelled
	// by the time the GitHub API round-trip completes.
	if err := session.Log(context.Background(), &mcp.LoggingMessageParams{
		Level:  "warning",
		Logger: "updater",
		Data:   msg,
	}); err != nil {
		log.Printf("[updater] failed to send MCP log notification: %v", err)
	}
}
