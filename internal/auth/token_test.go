package auth

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestConfigDir_Default(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")

	dir, err := configDir()
	require.NoError(t, err)

	home, err := os.UserHomeDir()
	require.NoError(t, err)

	expected := filepath.Join(home, ".config", appName)
	assert.Equal(t, expected, dir)
}

func TestConfigDir_XDG(t *testing.T) {
	xdgDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdgDir)

	dir, err := configDir()
	require.NoError(t, err)

	expected := filepath.Join(xdgDir, appName)
	assert.Equal(t, expected, dir)
}

func TestSaveAndLoadToken(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	original := &oauth2.Token{
		AccessToken:  "test-access-token",
		TokenType:    "Bearer",
		RefreshToken: "test-refresh-token",
		Expiry:       time.Now().Add(1 * time.Hour).Truncate(time.Second),
	}

	err := SaveToken(original)
	require.NoError(t, err)

	loaded, err := LoadToken()
	require.NoError(t, err)

	assert.Equal(t, original.AccessToken, loaded.AccessToken)
	assert.Equal(t, original.TokenType, loaded.TokenType)
	assert.Equal(t, original.RefreshToken, loaded.RefreshToken)
	assert.True(t, original.Expiry.Equal(loaded.Expiry), "expiry mismatch: want %v, got %v", original.Expiry, loaded.Expiry)
}

func TestLoadToken_NotFound(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	token, err := LoadToken()
	assert.Nil(t, token)
	assert.ErrorIs(t, err, ErrTokenNotFound)
}

func TestSaveToken_FilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	token := &oauth2.Token{
		AccessToken:  "test-access-token",
		TokenType:    "Bearer",
		RefreshToken: "test-refresh-token",
	}

	err := SaveToken(token)
	require.NoError(t, err)

	path := filepath.Join(tmpDir, appName, tokenFileName)
	info, err := os.Stat(path)
	require.NoError(t, err)

	assert.Equal(t, os.FileMode(filePerm), info.Mode().Perm())
}

func TestSaveToken_Nil(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	err := SaveToken(nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nil")
}
