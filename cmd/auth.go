package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ButeaLabs/butea-cli/internal/config"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
}

// ── auth logout ───────────────────────────────────────────────────────────────

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out and clear stored credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		cred, err := config.LoadCredentials()
		if err != nil {
			return err
		}
		if !cred.IsLoggedIn() {
			fmt.Println("Not currently logged in.")
			return nil
		}

		cfg, _, err := loadAll()
		if err != nil {
			return err
		}
		// Best-effort server-side revocation
		client := newClient(cfg, cred)
		_ = client.Logout(background(), cred.RefreshToken)

		if err := cred.Clear(); err != nil {
			return fmt.Errorf("clear credentials: %w", err)
		}
		fmt.Println("Logged out. Run 'butea init' to sign in again.")
		return nil
	},
}

// ── auth whoami ───────────────────────────────────────────────────────────────

var authWhoamiCmd = &cobra.Command{
	Use:     "whoami",
	Aliases: []string{"me"},
	Short:   "Print the currently authenticated user",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, cred, err := loadAll()
		if err != nil {
			return err
		}
		if err := cred.RequireAuth(); err != nil {
			return err
		}

		client := newClient(cfg, cred)
		user, err := client.GetMe(background())
		if err != nil {
			return fmt.Errorf("fetch profile: %w", err)
		}

		name := "<no name set>"
		if user.Name != nil && *user.Name != "" {
			name = *user.Name
		}

		fmt.Printf("ID:     %s\n", user.ID)
		fmt.Printf("Name:   %s\n", name)
		fmt.Printf("Email:  %s\n", user.Email)
		fmt.Printf("Active: %v\n", user.IsActive)
		if len(user.Providers) > 0 {
			providers := make([]string, len(user.Providers))
			for i, p := range user.Providers {
				providers[i] = p.Provider
			}
			fmt.Printf("OAuth:  %s\n", strings.Join(providers, ", "))
		}
		fmt.Printf("Since:  %s\n", user.CreatedAt.Format("Jan 2 2006"))
		return nil
	},
}

func init() {
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authWhoamiCmd)
	rootCmd.AddCommand(authCmd)
}

// ── terminal helpers (shared across cmd package) ──────────────────────────────

func prompt(label string) (string, error) {
	fmt.Print(label)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
