// Package cmd wires all Cobra commands for the butea CLI.
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ButeaLabs/butea-cli/internal/api"
	"github.com/ButeaLabs/butea-cli/internal/config"
)

// Version is the CLI version string, injected at build time via:
//
//	go build -ldflags "-X github.com/ButeaLabs/butea-cli/cmd.Version=v1.2.3"
var Version = "dev"

// SetVersion is called from main before Execute() so the build-time version
// string is visible to all commands and the --version flag works.
func SetVersion(v string) {
	Version = v
	rootCmd.Version = v
}

var (
	apiURLFlag string
	appURLFlag string
)

var rootCmd = &cobra.Command{
	Use:   "butea",
	Short: "butea CLI — deploy and manage your projects from the terminal",
	Long: `butea is the command-line interface for the Butea platform.

  butea init                    # authenticate & set up
  butea deploy                  # trigger a deployment
  butea projects list           # list your projects
  butea health                  # check API reachability

Full documentation: https://butea.in/docs/cli`,
}

// Execute is the entry-point called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&apiURLFlag, "api-url", "",
		"Backend API URL (overrides BUTEA_API_URL env and ~/.butea/config.toml)")
	rootCmd.PersistentFlags().StringVar(&appURLFlag, "app-url", "",
		"Frontend app URL (overrides BUTEA_APP_URL env and ~/.butea/config.toml)")
}

// ── shared helpers ────────────────────────────────────────────────────────────

// loadAll loads global config + credentials.
// Priority: CLI flags > env vars (applied in LoadGlobal) > config.toml > defaults.
func loadAll() (*config.GlobalConfig, *config.Credentials, error) {
	cfg, err := config.LoadGlobal()
	if err != nil {
		return nil, nil, fmt.Errorf("load config: %w", err)
	}
	// CLI flags are the highest priority override
	if apiURLFlag != "" {
		cfg.APIURL = apiURLFlag
	}
	if appURLFlag != "" {
		cfg.AppURL = appURLFlag
	}
	cred, err := config.LoadCredentials()
	if err != nil {
		return nil, nil, fmt.Errorf("load credentials: %w", err)
	}
	return cfg, cred, nil
}

// newClient builds an *api.Client and wires the auto-refresh persistence hook.
func newClient(cfg *config.GlobalConfig, cred *config.Credentials) *api.Client {
	c := api.NewClient(cfg.APIURL, cred.AccessToken, cred.RefreshToken, Version)
	c.OnTokenRefresh = func(access, refresh string) {
		cred.AccessToken = access
		cred.RefreshToken = refresh
		_ = cred.Save()
	}
	return c
}

// background returns a plain context for all CLI requests.
func background() context.Context { return context.Background() }

// fatal prints to stderr and exits 1.
func fatal(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}
