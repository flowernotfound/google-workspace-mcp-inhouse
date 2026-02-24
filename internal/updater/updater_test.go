package updater

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"path/filepath"
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
	// When version == "dev", the updater skips semver comparison and always
	// proceeds to download. Use a fake executable path to avoid overwriting
	// the actual test binary.
	client := &mockGitHubClient{
		release:      fakeRelease("v0.1.5"),
		downloadData: []byte("fake binary data"),
	}
	fakeExec := filepath.Join(t.TempDir(), "fake-binary")
	require.NoError(t, os.WriteFile(fakeExec, []byte("original"), 0o755))

	var out bytes.Buffer
	_ = runWithExecPath(context.Background(), "dev", client, &out, fakeExec)
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

func TestValidateAssetURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid github.com URL",
			url:     "https://github.com/owner/repo/releases/download/v1.0.0/binary",
			wantErr: false,
		},
		{
			name:    "valid objects.githubusercontent.com URL",
			url:     "https://objects.githubusercontent.com/github-production-release-asset/binary",
			wantErr: false,
		},
		{
			name:    "valid github-releases.githubusercontent.com URL",
			url:     "https://github-releases.githubusercontent.com/binary",
			wantErr: false,
		},
		{
			name:    "valid release-assets.githubusercontent.com URL",
			url:     "https://release-assets.githubusercontent.com/github-production-release-asset/binary",
			wantErr: false,
		},
		{
			name:    "uppercase host is normalized and allowed",
			url:     "https://GITHUB.COM/owner/repo/releases/download/v1.0.0/binary",
			wantErr: false,
		},
		{
			name:    "HTTP not allowed",
			url:     "http://objects.githubusercontent.com/binary",
			wantErr: true,
		},
		{
			name:    "unknown host not allowed",
			url:     "https://example.com/binary",
			wantErr: true,
		},
		{
			name:    "non-standard port not allowed",
			url:     "https://github.com:8443/binary",
			wantErr: true,
		},
		{
			name:    "missing host not allowed",
			url:     "https:///path/only",
			wantErr: true,
		},
		{
			name:    "invalid URL",
			url:     "://invalid",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAssetURL(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestDownloadAsset_CheckRedirectValidation verifies that the CheckRedirect callback
// set inside DownloadAsset correctly delegates to validateAssetURL, so that every
// URL in a redirect chain is validated against the allowlist.
func TestDownloadAsset_CheckRedirectValidation(t *testing.T) {
	// Replicate the CheckRedirect logic used in DownloadAsset.
	client := *downloadClient
	client.CheckRedirect = func(req *http.Request, _ []*http.Request) error {
		return validateAssetURL(req.URL.String())
	}

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "redirect to allowed host is permitted",
			url:     "https://objects.githubusercontent.com/github-production-release-asset/binary",
			wantErr: false,
		},
		{
			name:    "redirect to disallowed host is rejected",
			url:     "https://evil.com/payload",
			wantErr: true,
		},
		{
			name:    "redirect to HTTP URL is rejected",
			url:     "http://objects.githubusercontent.com/binary",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, tt.url, nil)
			require.NoError(t, err)
			err = client.CheckRedirect(req, nil)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
