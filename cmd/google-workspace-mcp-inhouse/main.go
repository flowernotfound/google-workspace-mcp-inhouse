package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/flowernotfound/google-workspace-mcp-inhouse/internal/auth"
)

func main() {
	log.SetOutput(os.Stderr)

	if len(os.Args) > 1 && os.Args[1] == "auth" {
		if err := auth.RunAuthFlow(context.Background()); err != nil {
			log.Fatalf("auth error: %v", err)
		}
		return
	}

	// TODO: MCP
	fmt.Fprintln(os.Stderr, "MCP server is not yet implemented. Coming in PR3.")
	os.Exit(1)
}
