package auth

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
)

const (
	appName       = "google-workspace-mcp-inhouse"
	tokenFileName = "token.json"
	dirPerm       = 0700
	filePerm      = 0600
)

// ErrTokenNotFound is returned when the token file does not exist.
var ErrTokenNotFound = errors.New("token file not found")

// configDir returns the application-specific config directory path.
// Uses $XDG_CONFIG_HOME if set, otherwise falls back to os.UserConfigDir().
func configDir() (string, error) {
	if base := os.Getenv("XDG_CONFIG_HOME"); base != "" {
		return filepath.Join(base, appName), nil
	}
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, appName), nil
}

// tokenPath returns the full path to the token file.
func tokenPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, tokenFileName), nil
}

// LoadToken reads the OAuth token from token.json.
// It returns ErrTokenNotFound if the file does not exist.
func LoadToken() (*oauth2.Token, error) {
	path, err := tokenPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrTokenNotFound
		}
		return nil, err
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}

	return &token, nil
}

// SaveToken writes the OAuth token to token.json.
// It creates the config directory if it does not exist.
// The file permission is set to 0600 (owner read/write only).
func SaveToken(token *oauth2.Token) error {
	if token == nil {
		return errors.New("token must not be nil")
	}

	dir, err := configDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, dirPerm); err != nil {
		return err
	}

	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(dir, tokenFileName)
	if err := os.WriteFile(path, data, filePerm); err != nil {
		return err
	}
	return os.Chmod(path, filePerm)
}
