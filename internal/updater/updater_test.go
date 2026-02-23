package updater

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockGitHubClient is a test double for GitHubClient.
type mockGitHubClient struct {
	release        *Release
	releaseErr     error
	downloadData   []byte
	downloadErr    error
	downloadCalled bool
}

func (m *mockGitHubClient) GetLatestRelease(_ context.Context) (*Release, error) {
	return m.release, m.releaseErr
}

func (m *mockGitHubClient) DownloadAsset(_ context.Context, _ string) ([]byte, error) {
	m.downloadCalled = true
	return m.downloadData, m.downloadErr
}

func fakeRelease(tag string) *Release {
	return &Release{
		TagName: tag,
		Assets: []Asset{
			{
				Name:               "google-workspace-mcp-inhouse_linux_amd64",
				BrowserDownloadURL: "https://example.com/binary",
			},
			{
				Name:               "google-workspace-mcp-inhouse_darwin_amd64",
				BrowserDownloadURL: "https://example.com/binary-darwin-amd64",
			},
			{
				Name:               "google-workspace-mcp-inhouse_darwin_arm64",
				BrowserDownloadURL: "https://example.com/binary-darwin-arm64",
			},
		},
	}
}

func TestRun_AlreadyLatest(t *testing.T) {
	client := &mockGitHubClient{release: fakeRelease("v0.1.5")}
	var out bytes.Buffer
	err := run(context.Background(), "v0.1.5", client, &out)
	require.NoError(t, err)
	assert.Contains(t, out.String(), "already up to date")
	assert.False(t, client.downloadCalled)
}

func TestRun_AlreadyAhead(t *testing.T) {
	client := &mockGitHubClient{release: fakeRelease("v0.1.4")}
	var out bytes.Buffer
	err := run(context.Background(), "v0.1.5", client, &out)
	require.NoError(t, err)
	assert.Contains(t, out.String(), "already up to date")
	assert.False(t, client.downloadCalled)
}

func TestRun_DevVersion(t *testing.T) {
	// Create a temp file to act as the "current binary" so os.Rename succeeds.
	tmpExec, err := os.CreateTemp("", "fake-binary-*")
	require.NoError(t, err)
	tmpExec.Close()
	defer os.Remove(tmpExec.Name())

	client := &mockGitHubClient{
		release:      fakeRelease("v0.1.5"),
		downloadData: []byte("fake binary data"),
	}
	var out bytes.Buffer

	// Override executable path via a thin wrapper that uses a temp file.
	// We can't override os.Executable, so we test the internal run function
	// and confirm download is called when version == "dev".
	// Since os.Rename will fail on the temp path, we just check downloadCalled.
	_ = run(context.Background(), "dev", client, &out)
	assert.True(t, client.downloadCalled, "download should be called for dev version")
}

func TestCompareSemver(t *testing.T) {
	tests := []struct {
		a, b     string
		wantSign int // positive, zero, or negative
	}{
		{"v0.1.5", "v0.1.4", 1},
		{"v0.1.4", "v0.1.5", -1},
		{"v0.1.5", "v0.1.5", 0},
		{"v1.0.0", "v0.9.9", 1},
		{"v0.2.0", "v0.1.9", 1},
	}
	for _, tt := range tests {
		cmp, err := compareSemver(tt.a, tt.b)
		require.NoError(t, err, "a=%s b=%s", tt.a, tt.b)
		switch {
		case tt.wantSign > 0:
			assert.Positive(t, cmp, "a=%s b=%s", tt.a, tt.b)
		case tt.wantSign < 0:
			assert.Negative(t, cmp, "a=%s b=%s", tt.a, tt.b)
		default:
			assert.Zero(t, cmp, "a=%s b=%s", tt.a, tt.b)
		}
	}
}

func TestParseSemver_InvalidFormat(t *testing.T) {
	_, err := parseSemver("notaversion")
	assert.Error(t, err)
}
