package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"gigtape/adapters/spotify"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

// tokenFilePath is the on-disk location of the cached OAuth token.
func tokenFilePath() string {
	return filepath.Join(os.TempDir(), "gigtape-token.json")
}

type cachedToken struct {
	Token     *oauth2.Token `json:"token"`
	UserID    string        `json:"user_id"`
	ClientID  string        `json:"client_id"`
}

func saveToken(t cachedToken) error {
	b, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(tokenFilePath(), b, 0600)
}

func loadToken() (cachedToken, error) {
	b, err := os.ReadFile(tokenFilePath())
	if err != nil {
		return cachedToken{}, err
	}
	var t cachedToken
	if err := json.Unmarshal(b, &t); err != nil {
		return cachedToken{}, err
	}
	return t, nil
}

func authCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with Spotify via OAuth PKCE",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAuth(cmd.Context())
		},
	}
}

func runAuth(ctx context.Context) error {
	clientID := deps.SpotifyClientID
	if clientID == "" {
		return errors.New("SPOTIFY_CLIENT_ID is not set")
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	defer listener.Close()
	redirectURI := fmt.Sprintf("http://127.0.0.1:%d/callback", listener.Addr().(*net.TCPAddr).Port)

	challenge, err := spotify.GenerateChallenge()
	if err != nil {
		return err
	}
	state := fmt.Sprintf("%d", time.Now().UnixNano())

	authURL := spotify.AuthURL(clientID, redirectURI, challenge.Challenge, state)
	fmt.Println("Opening browser for Spotify authorization…")
	if err := openBrowser(authURL); err != nil {
		fmt.Printf("Please open this URL manually:\n%s\n", authURL)
	}

	type result struct {
		code string
		err  error
	}
	done := make(chan result, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		gotState := r.URL.Query().Get("state")
		code := r.URL.Query().Get("code")
		if gotState != state || code == "" {
			http.Error(w, "OAuth handshake failed.", http.StatusBadRequest)
			done <- result{err: errors.New("state mismatch or missing code")}
			return
		}
		fmt.Fprintln(w, "Authorization received. You can close this tab.")
		done <- result{code: code}
	})
	srv := &http.Server{Handler: mux}
	go func() { _ = srv.Serve(listener) }()
	defer srv.Shutdown(context.Background())

	select {
	case <-ctx.Done():
		return ctx.Err()
	case res := <-done:
		if res.err != nil {
			return res.err
		}
		token, err := spotify.ExchangeCode(ctx, clientID, redirectURI, res.code, challenge.Verifier)
		if err != nil {
			return err
		}
		httpClient := spotify.NewClient(ctx, token, clientID)
		userID, err := spotify.GetCurrentUserID(ctx, httpClient)
		if err != nil {
			return err
		}
		if err := saveToken(cachedToken{Token: token, UserID: userID, ClientID: clientID}); err != nil {
			return err
		}
		fmt.Printf("Authenticated as %s\n", userID)
		return nil
	}
}

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
		return errors.New("unsupported platform")
	}
	return cmd.Start()
}
