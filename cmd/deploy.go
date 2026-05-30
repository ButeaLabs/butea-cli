package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ButeaLabs/butea-cli/internal/config"
)

var deployCmd = &cobra.Command{
	Use:   "deploy [projectId]",
	Short: "Trigger a new deployment",
	Long: `Trigger a deployment for a project.

If run from a directory containing .butea.toml (created by 'butea init'),
the project ID and branch are read automatically.

Flags always override .butea.toml.`,
	Example: `  butea deploy                          # uses .butea.toml
  butea deploy abc-123                  # explicit project ID
  butea deploy abc-123 --branch staging`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, cred, err := loadAll()
		if err != nil {
			return err
		}
		if err := cred.RequireAuth(); err != nil {
			return err
		}

		branchFlag, _ := cmd.Flags().GetString("branch")
		projectID := ""
		branch := branchFlag

		if len(args) > 0 {
			projectID = args[0]
		}

		// Read .butea.toml if no explicit project ID
		if projectID == "" {
			lc, lcErr := config.LoadLocal()
			if lcErr == nil && lc != nil {
				projectID = lc.ProjectID
				if branch == "" {
					branch = lc.Branch
				}
			}
		}

		if projectID == "" {
			return fmt.Errorf("project ID required – pass it as an argument or run 'butea init' to link a project")
		}

		client := newClient(cfg, cred)

		// Auto-resolve default branch if still empty
		if branch == "" {
			p, pErr := client.GetProject(background(), projectID)
			if pErr != nil {
				return fmt.Errorf("get project (resolve branch): %w", pErr)
			}
			branch = p.DefaultBranch
		}

		fmt.Printf("Deploying project %s on branch '%s'…\n", projectID, branch)

		dep, err := client.CreateDeployment(background(), projectID, branch)
		if err != nil {
			return fmt.Errorf("create deployment: %w", err)
		}

		fmt.Printf("✓ Deployment created\n\n")
		fmt.Printf("  ID:      %s\n", dep.ID)
		fmt.Printf("  Branch:  %s\n", dep.Branch)
		fmt.Printf("  Status:  %s\n", dep.Status)
		if dep.DeployURL != "" {
			fmt.Printf("  URL:     %s\n", dep.DeployURL)
		}
		fmt.Printf("  Created: %s\n", dep.CreatedAt.Format("Jan 2, 2006 15:04 MST"))
		fmt.Printf("\nTrack: butea deployments get %s\n", dep.ID)
		return nil
	},
}

func init() {
	deployCmd.Flags().StringP("branch", "b", "", "Branch to deploy (defaults to project default)")
	rootCmd.AddCommand(deployCmd)
}

