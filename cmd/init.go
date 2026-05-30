package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ButeaLabs/butea-cli/internal/auth"
	"github.com/ButeaLabs/butea-cli/internal/config"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Authenticate and set up butea in the current project",
	Long: `init authenticates the CLI via your browser and creates:

  ~/.butea/
  ├── config.toml   global settings (API URL, app URL)
  └── cred.toml     credentials    (access + refresh tokens)

If run inside a project repository it also creates:

  .butea.toml       links this directory to a Butea project

The .butea.toml file lets you run 'butea deploy' without flags.
Add it to .gitignore or commit it to share with your team.`,
	Example: `  butea init              # authenticate + optionally link a project
  butea init --reauth     # force re-authentication even if already logged in`,
	RunE: func(cmd *cobra.Command, args []string) error {
		reauth, _ := cmd.Flags().GetBool("reauth")
		linkProject, _ := cmd.Flags().GetBool("link")

		// ── Load existing config (or defaults) ────────────────────────────
		cfg, err := config.LoadGlobal()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		if apiURLFlag != "" {
			cfg.APIURL = apiURLFlag
		}
		if appURLFlag != "" {
			cfg.AppURL = appURLFlag
		}

		cred, err := config.LoadCredentials()
		if err != nil {
			return fmt.Errorf("load credentials: %w", err)
		}

		// ── Step 1: Authenticate ───────────────────────────────────────────
		if cred.IsLoggedIn() && !reauth {
			client := newClient(cfg, cred)
			user, err := client.GetMe(background())
			if err == nil {
				name := user.Email
				if user.Name != nil && *user.Name != "" {
					name = *user.Name
				}
				fmt.Printf("✓ Already authenticated as %s\n", name)
				fmt.Printf("  Run 'butea init --reauth' to sign in with a different account.\n\n")
				goto linkStep
			}
			// Token invalid — fall through to re-auth
			fmt.Println("  Stored token is invalid. Re-authenticating…")
		}

		{
			fmt.Printf("\n  butea CLI setup\n")
			fmt.Printf("  %s\n\n", strings.Repeat("─", 38))

			result, err := auth.StartBrowserFlow(background(), cfg.AppURL, 5*time.Minute)
			if err != nil {
				return fmt.Errorf("authentication failed: %w", err)
			}

			cred.AccessToken = result.AccessToken
			cred.RefreshToken = result.RefreshToken

			// Save global config + credentials
			if saveErr := cfg.Save(); saveErr != nil {
				return fmt.Errorf("save config: %w", saveErr)
			}
			if saveErr := cred.Save(); saveErr != nil {
				return fmt.Errorf("save credentials: %w", saveErr)
			}

			// Confirm identity
			client := newClient(cfg, cred)
			user, err := client.GetMe(background())
			if err != nil {
				return fmt.Errorf("verify authentication: %w", err)
			}
			name := user.Email
			if user.Name != nil && *user.Name != "" {
				name = *user.Name
			}
			fmt.Printf("\n  ✓ Authenticated as %s (%s)\n", name, user.Email)

			dir, _ := config.GlobalDir()
			fmt.Printf("  ✓ Credentials saved to %s\n\n", dir)
		}

	linkStep:
		// ── Step 2: Optionally link current directory to a project ─────────
		if !linkProject {
			answer, promptErr := prompt("Link this directory to a Butea project? [y/N]: ")
			if promptErr != nil || (answer != "y" && answer != "Y" && answer != "yes") {
				fmt.Println("\nDone! Run 'butea projects list' to see your projects.")
				fmt.Println("Run 'butea deploy <projectId>' to trigger a deployment.")
				return nil
			}
		}

		cfg2, cred2, err := loadAll()
		if err != nil {
			return err
		}
		client := newClient(cfg2, cred2)

		resp, err := client.ListProjects(background())
		if err != nil {
			return fmt.Errorf("list projects: %w", err)
		}
		if len(resp.Projects) == 0 {
			fmt.Println("\nNo projects found. Import one at", cfg2.AppURL+"/import")
			return nil
		}

		fmt.Println("\n  Your projects:")
		for i, p := range resp.Projects {
			fmt.Printf("  [%d] %s  (%s)\n", i+1, p.Name, p.RepoFullName)
		}
		fmt.Println()

		choice, err := prompt("Enter number (or project ID): ")
		if err != nil {
			return err
		}

		var projectID string
		// Check if input is a numeric index
		var idx int
		if _, err := fmt.Sscanf(choice, "%d", &idx); err == nil && idx >= 1 && idx <= len(resp.Projects) {
			projectID = resp.Projects[idx-1].ID
		} else {
			projectID = strings.TrimSpace(choice)
		}

		project, err := client.GetProject(background(), projectID)
		if err != nil {
			return fmt.Errorf("fetch project: %w", err)
		}
		if project == nil {
			return fmt.Errorf("project not found")
		}

		lc := &config.LocalConfig{
			ProjectID: project.ID,
			Branch:    project.DefaultBranch,
		}
		if err := lc.Save(); err != nil {
			return fmt.Errorf("write .butea.toml: %w", err)
		}

		fmt.Printf("\n  ✓ Linked to project '%s'\n", project.Name)
		fmt.Printf("  ✓ .butea.toml created\n\n")
		fmt.Printf("  project_id = %q\n", project.ID)
		fmt.Printf("  branch     = %q\n\n", project.DefaultBranch)
		fmt.Println("  Run 'butea deploy' from this directory to trigger a deployment.")
		fmt.Println("  Add .butea.toml to .gitignore or commit it to share with your team.")
		return nil
	},
}

func init() {
	initCmd.Flags().Bool("reauth", false, "Force re-authentication even if already logged in")
	initCmd.Flags().Bool("link", false, "Skip the link prompt and always link a project")
	rootCmd.AddCommand(initCmd)
}
