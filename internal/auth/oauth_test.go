package auth

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2/google"
)

const testCredentialsJSON = `{
  "installed": {
    "client_id": "test-id.apps.googleusercontent.com",
    "client_secret": "test-secret",
    "redirect_uris": ["http://localhost"],
    "auth_uri": "https://accounts.google.com/o/oauth2/auth",
    "token_uri": "https://oauth2.googleapis.com/token"
  }
}`

func TestLoadClientCredentials_EnvVars(t *testing.T) {
	t.Setenv("GOOGLE_CLIENT_ID", "env-client-id")
	t.Setenv("GOOGLE_CLIENT_SECRET", "env-client-secret")
	t.Setenv("GOOGLE_CREDENTIALS_FILE", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	clientID, clientSecret, err := loadClientCredentials()
	require.NoError(t, err)
	assert.Equal(t, "env-client-id", clientID)
	assert.Equal(t, "env-client-secret", clientSecret)
}

func TestLoadClientCredentials_EnvVarsPartial(t *testing.T) {
	t.Setenv("GOOGLE_CLIENT_SECRET", "")
	t.Setenv("GOOGLE_CREDENTIALS_FILE", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	t.Run("ID only", func(t *testing.T) {
		t.Setenv("GOOGLE_CLIENT_ID", "only-id")
		t.Setenv("GOOGLE_CLIENT_SECRET", "")

		_, _, err := loadClientCredentials()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "both GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET must be set")
	})

	t.Run("Secret only", func(t *testing.T) {
		t.Setenv("GOOGLE_CLIENT_ID", "")
		t.Setenv("GOOGLE_CLIENT_SECRET", "only-secret")

		_, _, err := loadClientCredentials()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "both GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET must be set")
	})
}

func TestLoadClientCredentials_CredentialsFile(t *testing.T) {
	t.Setenv("GOOGLE_CLIENT_ID", "")
	t.Setenv("GOOGLE_CLIENT_SECRET", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	tmpFile := filepath.Join(t.TempDir(), "credentials.json")
	err := os.WriteFile(tmpFile, []byte(testCredentialsJSON), 0600)
	require.NoError(t, err)

	t.Setenv("GOOGLE_CREDENTIALS_FILE", tmpFile)

	clientID, clientSecret, err := loadClientCredentials()
	require.NoError(t, err)
	assert.Equal(t, "test-id.apps.googleusercontent.com", clientID)
	assert.Equal(t, "test-secret", clientSecret)
}

func TestLoadClientCredentials_DefaultPath(t *testing.T) {
	t.Setenv("GOOGLE_CLIENT_ID", "")
	t.Setenv("GOOGLE_CLIENT_SECRET", "")
	t.Setenv("GOOGLE_CREDENTIALS_FILE", "")

	xdgDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdgDir)

	credDir := filepath.Join(xdgDir, appName)
	err := os.MkdirAll(credDir, 0700)
	require.NoError(t, err)

	credPath := filepath.Join(credDir, credentialsFileName)
	err = os.WriteFile(credPath, []byte(testCredentialsJSON), 0600)
	require.NoError(t, err)

	clientID, clientSecret, err := loadClientCredentials()
	require.NoError(t, err)
	assert.Equal(t, "test-id.apps.googleusercontent.com", clientID)
	assert.Equal(t, "test-secret", clientSecret)
}

func TestLoadClientCredentials_NotFound(t *testing.T) {
	t.Setenv("GOOGLE_CLIENT_ID", "")
	t.Setenv("GOOGLE_CLIENT_SECRET", "")
	t.Setenv("GOOGLE_CREDENTIALS_FILE", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	_, _, err := loadClientCredentials()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "credentials not found")
}

func TestLoadClientCredentials_InvalidJSON(t *testing.T) {
	t.Setenv("GOOGLE_CLIENT_ID", "")
	t.Setenv("GOOGLE_CLIENT_SECRET", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	tmpFile := filepath.Join(t.TempDir(), "bad.json")
	err := os.WriteFile(tmpFile, []byte("not valid json"), 0600)
	require.NoError(t, err)

	t.Setenv("GOOGLE_CREDENTIALS_FILE", tmpFile)

	_, _, err = loadClientCredentials()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid credentials file format")
}

func TestLoadClientCredentials_MissingInstalledKey(t *testing.T) {
	t.Setenv("GOOGLE_CLIENT_ID", "")
	t.Setenv("GOOGLE_CLIENT_SECRET", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	tmpFile := filepath.Join(t.TempDir(), "empty.json")
	err := os.WriteFile(tmpFile, []byte(`{"web": {}}`), 0600)
	require.NoError(t, err)

	t.Setenv("GOOGLE_CREDENTIALS_FILE", tmpFile)

	_, _, err = loadClientCredentials()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "'installed' key not found")
}

func TestNewOAuthConfig(t *testing.T) {
	config := newOAuthConfig("test-id", "test-secret", "http://localhost:8080")

	assert.Equal(t, "test-id", config.ClientID)
	assert.Equal(t, "test-secret", config.ClientSecret)
	assert.Equal(t, "http://localhost:8080", config.RedirectURL)
	assert.Equal(t, scopes, config.Scopes)
	assert.Equal(t, google.Endpoint, config.Endpoint)
}

func TestAuthorize_NoToken(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("GOOGLE_CLIENT_ID", "")
	t.Setenv("GOOGLE_CLIENT_SECRET", "")
	t.Setenv("GOOGLE_CREDENTIALS_FILE", "")

	client, err := Authorize(context.Background())
	assert.Nil(t, client)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "auth")
}

func TestLoadClientCredentials_EmptyFields(t *testing.T) {
	t.Setenv("GOOGLE_CLIENT_ID", "")
	t.Setenv("GOOGLE_CLIENT_SECRET", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	emptyFieldsJSON := `{
  "installed": {
    "client_id": "",
    "client_secret": "",
    "redirect_uris": ["http://localhost"]
  }
}`
	tmpFile := filepath.Join(t.TempDir(), "credentials.json")
	err := os.WriteFile(tmpFile, []byte(emptyFieldsJSON), 0600)
	require.NoError(t, err)

	t.Setenv("GOOGLE_CREDENTIALS_FILE", tmpFile)

	_, _, err = loadClientCredentials()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty")
}
