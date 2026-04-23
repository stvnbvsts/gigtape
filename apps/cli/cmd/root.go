// Package cmd wires Cobra commands for the gigtape CLI.
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var envFile string

// Deps holds collaborators injected from main.go. Each command reads from this
// struct so the composition root stays in main.
type Deps struct {
	SetlistfmAPIKey     string
	SpotifyClientID     string
	SpotifyRedirectURI  string
}

var deps Deps

// Root returns the root Cobra command configured with the given dependencies.
func Root(d Deps) *cobra.Command {
	deps = d
	root := &cobra.Command{
		Use:   "gigtape",
		Short: "Gigtape — create Spotify playlists from setlist.fm data",
	}
	root.PersistentFlags().StringVar(&envFile, "env-file", ".env", "Path to .env file with API keys")
	root.AddCommand(authCmd())
	root.AddCommand(artistCmd())
	return root
}

// Execute runs the root command; called by main.
func Execute(d Deps) error {
	return Root(d).Execute()
}

// prompt reads a single line of input from stdin.
func prompt(msg string) string {
	fmt.Print(msg)
	r := bufio.NewReader(os.Stdin)
	line, _ := r.ReadString('\n')
	return strings.TrimSpace(line)
}

// confirm prompts with [y/n] and returns true on y/yes.
func confirm(msg string) bool {
	ans := strings.ToLower(prompt(msg + " [y/n]: "))
	return ans == "y" || ans == "yes"
}
