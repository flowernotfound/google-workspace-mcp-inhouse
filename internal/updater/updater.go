package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
)

const (
	owner = "flowernotfound"
	repo  = "google-workspace-mcp-inhouse"
)

// GitHubClient abstracts GitHub API access for testability.
type GitHubClient interface {
	GetLatestRelease(ctx context.Context) (*Release, error)
	DownloadAsset(ctx context.Context, url string) ([]byte, error)
}

// Release represents a GitHub release.
type Release struct {
	TagName string
	Assets  []Asset
}

// Asset represents a binary asset attached to a release.
type Asset struct {
	Name               string
	BrowserDownloadURL string
}

// Run checks for updates and replaces the binary if a newer version is available.
// currentVersion should be the value injected by ldflags (e.g. "v0.1.42" or "dev").
func Run(ctx context.Context, currentVersion string) error {
	return run(ctx, currentVersion, &httpGitHubClient{}, os.Stdout)
}

func run(ctx context.Context, currentVersion string, client GitHubClient, out io.Writer) error {
	release, err := client.GetLatestRelease(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch latest release: %w", err)
	}

	latestVersion := release.TagName

	if currentVersion != "dev" {
		cmp, err := compareSemver(latestVersion, currentVersion)
		if err != nil {
			return fmt.Errorf("failed to compare versions: %w", err)
		}
		if cmp <= 0 {
			fmt.Fprintf(out, "already up to date (%s)\n", currentVersion)
			return nil
		}
	}

	if runtime.GOOS == "windows" {
		fmt.Fprintf(out, "To update on Windows, re-run install.ps1:\n")
		fmt.Fprintf(out, "  irm https://raw.githubusercontent.com/%s/%s/master/install.ps1 | iex\n", owner, repo)
		return nil
	}

	assetName := assetNameForPlatform()
	assetURL := ""
	for _, a := range release.Assets {
		if a.Name == assetName {
			assetURL = a.BrowserDownloadURL
			break
		}
	}
	if assetURL == "" {
		return fmt.Errorf("no asset found for platform: %s", assetName)
	}

	fmt.Fprintf(out, "updating %s -> %s\n", currentVersion, latestVersion)

	data, err := client.DownloadAsset(ctx, assetURL)
	if err != nil {
		return fmt.Errorf("failed to download asset: %w", err)
	}

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	tmpFile, err := os.CreateTemp("", "google-workspace-mcp-inhouse-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()

	if err := os.Chmod(tmpPath, 0o755); err != nil {
		return fmt.Errorf("failed to chmod temp file: %w", err)
	}

	if err := os.Rename(tmpPath, execPath); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	fmt.Fprintf(out, "updated to %s\n", latestVersion)
	return nil
}

// assetNameForPlatform returns the expected asset name for the current OS/arch.
func assetNameForPlatform() string {
	goarch := runtime.GOARCH
	return fmt.Sprintf("google-workspace-mcp-inhouse_%s_%s", runtime.GOOS, goarch)
}

// compareSemver returns positive if a > b, 0 if equal, negative if a < b.
// It expects versions in the form "vX.Y.Z" or "X.Y.Z".
func compareSemver(a, b string) (int, error) {
	pa, err := parseSemver(a)
	if err != nil {
		return 0, fmt.Errorf("invalid version %q: %w", a, err)
	}
	pb, err := parseSemver(b)
	if err != nil {
		return 0, fmt.Errorf("invalid version %q: %w", b, err)
	}
	for i := 0; i < 3; i++ {
		if pa[i] != pb[i] {
			return pa[i] - pb[i], nil
		}
	}
	return 0, nil
}

// parseSemver parses "vX.Y.Z" or "X.Y.Z" into [3]int.
func parseSemver(v string) ([3]int, error) {
	v = strings.TrimPrefix(v, "v")
	parts := strings.SplitN(v, ".", 3)
	if len(parts) != 3 {
		return [3]int{}, fmt.Errorf("expected X.Y.Z format")
	}
	var result [3]int
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return [3]int{}, err
		}
		result[i] = n
	}
	return result, nil
}

// httpGitHubClient is the real implementation using net/http.
type httpGitHubClient struct{}

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func (c *httpGitHubClient) GetLatestRelease(ctx context.Context) (*Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var gr githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&gr); err != nil {
		return nil, err
	}

	release := &Release{TagName: gr.TagName}
	for _, a := range gr.Assets {
		release.Assets = append(release.Assets, Asset(a))
	}
	return release, nil
}

func (c *httpGitHubClient) DownloadAsset(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
