package google

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/sheets/v4"
)

// Scopes returns the OAuth scopes required by this application
func Scopes() []string {
	return []string{
		sheets.SpreadsheetsScope,
		drive.DriveMetadataReadonlyScope,
	}
}

// Authenticator provides HTTP client for Google API authentication
type Authenticator interface {
	GetClient(ctx context.Context) (*http.Client, error)
}

// OAuthAuthenticator implements Authenticator using OAuth2
type OAuthAuthenticator struct {
	credentialsFile string
	tokenFile       string
}

// NewOAuthAuthenticator creates a new OAuthAuthenticator
func NewOAuthAuthenticator(credentialsFile, tokenFile string) *OAuthAuthenticator {
	return &OAuthAuthenticator{
		credentialsFile: credentialsFile,
		tokenFile:       tokenFile,
	}
}

// GetClient returns an authenticated HTTP client using OAuth2
func (a *OAuthAuthenticator) GetClient(ctx context.Context) (*http.Client, error) {
	b, err := os.ReadFile(a.credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, Scopes()...)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}

	token, err := a.tokenFromFile()
	if err != nil {
		return nil, fmt.Errorf("token not found, please run 'gsheet auth' first: %v", err)
	}

	tokenSource := config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("unable to refresh token, please run 'gsheet auth' again: %v", err)
	}

	if newToken.AccessToken != token.AccessToken {
		if err := a.saveToken(newToken); err != nil {
			return nil, fmt.Errorf("unable to save refreshed token: %v", err)
		}
	}

	return oauth2.NewClient(ctx, tokenSource), nil
}

func (a *OAuthAuthenticator) tokenFromFile() (*oauth2.Token, error) {
	f, err := os.Open(a.tokenFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	return token, err
}

func (a *OAuthAuthenticator) saveToken(token *oauth2.Token) error {
	f, err := os.Create(a.tokenFile)
	if err != nil {
		return fmt.Errorf("unable to cache oauth token: %v", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}

// Authenticate runs the OAuth flow with local server callback and saves the token
func (a *OAuthAuthenticator) Authenticate() error {
	b, err := os.ReadFile(a.credentialsFile)
	if err != nil {
		return fmt.Errorf("unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, Scopes()...)
	if err != nil {
		return fmt.Errorf("unable to parse client secret file to config: %v", err)
	}

	// Find available port
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return fmt.Errorf("unable to start local server: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	redirectURL := fmt.Sprintf("http://localhost:%d/callback", port)

	// Override redirect URL
	config.RedirectURL = redirectURL

	// Channel to receive the authorization code
	codeChan := make(chan string)
	errChan := make(chan error)

	// Start local server
	server := &http.Server{}
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			errChan <- fmt.Errorf("no code in callback")
			http.Error(w, "No code received", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<html><body><h1>Authentication successful!</h1><p>You can close this window.</p></body></html>`)
		codeChan <- code
	})

	go func() {
		if err := server.Serve(listener); err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Generate auth URL
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	fmt.Printf("Opening browser for authentication...\n")
	fmt.Printf("If browser doesn't open, visit this URL:\n%s\n", authURL)

	// Open browser
	openBrowser(authURL)

	// Wait for callback
	var code string
	select {
	case code = <-codeChan:
	case err := <-errChan:
		server.Close()
		return fmt.Errorf("authentication failed: %v", err)
	}

	// Shutdown server
	server.Close()

	// Exchange code for token
	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		return fmt.Errorf("unable to retrieve token: %v", err)
	}

	return a.saveToken(token)
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "darwin":
		err = exec.Command("open", url).Start()
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	}
	if err != nil {
		fmt.Printf("Failed to open browser: %v\n", err)
	}
}

// ServiceAccountAuthenticator implements Authenticator using Service Account
type ServiceAccountAuthenticator struct {
	credentialsFile string
}

// NewServiceAccountAuthenticator creates a new ServiceAccountAuthenticator
func NewServiceAccountAuthenticator(credentialsFile string) *ServiceAccountAuthenticator {
	return &ServiceAccountAuthenticator{
		credentialsFile: credentialsFile,
	}
}

// GetClient returns an authenticated HTTP client using Service Account
func (a *ServiceAccountAuthenticator) GetClient(ctx context.Context) (*http.Client, error) {
	if err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", a.credentialsFile); err != nil {
		return nil, fmt.Errorf("unable to set GOOGLE_APPLICATION_CREDENTIALS: %v", err)
	}
	// Return nil to use Application Default Credentials
	return nil, nil
}
