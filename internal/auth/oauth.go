package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var scopes = []string{
	"https://www.googleapis.com/auth/documents.readonly",
	"https://www.googleapis.com/auth/drive.readonly",
	"https://www.googleapis.com/auth/spreadsheets.readonly",
}

const credentialsFileName = "credentials.json"

// credentialsFile represents the structure of credentials.json downloaded from Google Cloud Console.
type credentialsFile struct {
	Installed *credentialsData `json:"installed"`
}

type credentialsData struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURIs []string `json:"redirect_uris"`
	AuthURI      string   `json:"auth_uri"`
	TokenURI     string   `json:"token_uri"`
}

// loadClientCredentials resolves client credentials in the following priority:
//  1. GOOGLE_CLIENT_ID + GOOGLE_CLIENT_SECRET environment variables
//  2. File specified by GOOGLE_CREDENTIALS_FILE environment variable
//  3. Default path: $XDG_CONFIG_HOME/google-workspace-mcp-inhouse/credentials.json
func loadClientCredentials() (clientID, clientSecret string, err error) {
	// 1. Environment variables
	clientID = os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
	if clientID != "" && clientSecret != "" {
		return clientID, clientSecret, nil
	}
	if clientID != "" || clientSecret != "" {
		return "", "", errors.New("both GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET must be set")
	}

	// 2. GOOGLE_CREDENTIALS_FILE
	if path := os.Getenv("GOOGLE_CREDENTIALS_FILE"); path != "" {
		return loadCredentialsFromFile(path)
	}

	// 3. Default path
	dir, err := configDir()
	if err != nil {
		return "", "", err
	}
	defaultPath := filepath.Join(dir, credentialsFileName)
	id, secret, err := loadCredentialsFromFile(defaultPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", "", fmt.Errorf(
				"credentials not found. Please configure one of the following:\n"+
					"  1. Set GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET environment variables\n"+
					"  2. Set GOOGLE_CREDENTIALS_FILE environment variable to the file path\n"+
					"  3. Place credentials.json at %s", defaultPath)
		}
		return "", "", err
	}
	return id, secret, nil
}

// loadCredentialsFromFile reads and parses a credentials.json file.
func loadCredentialsFromFile(path string) (clientID, clientSecret string, err error) {
	data, err := os.ReadFile(path) //nolint:gosec // G703: path may be user-controlled via GOOGLE_CREDENTIALS_FILE, but this is intentional to allow custom credential locations in a local CLI tool
	if err != nil {
		return "", "", fmt.Errorf("failed to read credentials file %q: %w", path, err)
	}

	var creds credentialsFile
	if err := json.Unmarshal(data, &creds); err != nil {
		return "", "", fmt.Errorf("invalid credentials file format: %w", err)
	}

	if creds.Installed == nil {
		return "", "", errors.New("'installed' key not found in credentials file")
	}

	if creds.Installed.ClientID == "" || creds.Installed.ClientSecret == "" {
		return "", "", fmt.Errorf("client_id or client_secret is empty in credentials file %q", path)
	}

	return creds.Installed.ClientID, creds.Installed.ClientSecret, nil
}

// newOAuthConfig creates an oauth2.Config with the given credentials and redirect URL.
func newOAuthConfig(clientID, clientSecret, redirectURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       scopes,
		Endpoint:     google.Endpoint,
		RedirectURL:  redirectURL,
	}
}

// generateState generates a random state parameter for CSRF protection.
func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// RunAuthFlow runs the interactive OAuth 2.0 authentication flow.
// It opens a browser for Google login, receives the authorization code via a local
// callback server, exchanges it for tokens, and saves them to token.json.
func RunAuthFlow(ctx context.Context) error {
	clientID, clientSecret, err := loadClientCredentials()
	if err != nil {
		return err
	}

	// Start listener on a random port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("failed to start callback server: %w", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	redirectURL := fmt.Sprintf("http://127.0.0.1:%d/callback", port)

	config := newOAuthConfig(clientID, clientSecret, redirectURL)

	state, err := generateState()
	if err != nil {
		return fmt.Errorf("failed to generate state parameter: %w", err)
	}

	authURL := config.AuthCodeURL(state, oauth2.AccessTypeOffline)

	log.Printf("Open the following URL in your browser to authenticate:\n%s\n", authURL)

	if err := openBrowser(authURL); err != nil {
		log.Printf("Failed to open browser automatically: %v. Please open the URL above manually.", err)
	}

	// Wait for callback
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	sendCode := func(code string) {
		select {
		case codeCh <- code:
		default:
		}
	}
	sendErr := func(err error) {
		select {
		case errCh <- err:
		default:
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			http.Error(w, "invalid request", http.StatusBadRequest)
			sendErr(errors.New("state parameter mismatch (possible CSRF attack)"))
			return
		}

		if errMsg := r.URL.Query().Get("error"); errMsg != "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprintf(w, "<h1>Authentication failed</h1><p>%s</p>", html.EscapeString(errMsg))
			sendErr(fmt.Errorf("authentication denied: %s", errMsg))
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "authorization code not found", http.StatusBadRequest)
			sendErr(errors.New("callback did not contain an authorization code"))
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, "<h1>Authentication successful</h1><p>You can close this tab.</p>")
		sendCode(code)
	})

	server := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			sendErr(fmt.Errorf("callback server error: %w", err))
		}
	}()

	shutdown := func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("warning: callback server shutdown error: %v", err)
		}
	}

	var code string
	select {
	case code = <-codeCh:
		// Authorization code received
	case err := <-errCh:
		shutdown()
		return err
	case <-ctx.Done():
		shutdown()
		return ctx.Err()
	}

	shutdown()

	// Exchange authorization code for token
	token, err := config.Exchange(ctx, code)
	if err != nil {
		return fmt.Errorf("failed to exchange token: %w", err)
	}

	if err := SaveToken(token); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	log.Println("Authentication successful. Token saved.")
	return nil
}

// Authorize returns an authenticated http.Client using the saved token.
// It returns an error prompting the user to run the auth subcommand if no token exists.
// Expired tokens are automatically refreshed using the refresh_token.
func Authorize(ctx context.Context) (*http.Client, error) {
	token, err := LoadToken()
	if err != nil {
		if errors.Is(err, ErrTokenNotFound) {
			return nil, fmt.Errorf("authentication required. Run '%s auth' first", os.Args[0])
		}
		return nil, fmt.Errorf("failed to load token: %w", err)
	}

	clientID, clientSecret, err := loadClientCredentials()
	if err != nil {
		return nil, err
	}

	config := newOAuthConfig(clientID, clientSecret, "")
	tokenSource := config.TokenSource(ctx, token)

	// Save token if refreshed
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token. Run '%s auth' to re-authenticate: %w", os.Args[0], err)
	}

	if newToken.AccessToken != token.AccessToken {
		if err := SaveToken(newToken); err != nil {
			log.Printf("warning: failed to save refreshed token: %v", err)
		}
	}

	return oauth2.NewClient(ctx, tokenSource), nil
}

// openBrowser opens the given URL in the default browser.
func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
	return cmd.Start()
}
