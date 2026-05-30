package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var deploymentsCmd = &cobra.Command{
	Use:   "deployments",
	Short: "View and manage deployments",
}

var deploymentsListCmd = &cobra.Command{
	Use:     "list <projectId>",
	Aliases: []string{"ls"},
	Short:   "List deployments for a project",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, cred, err := loadAll()
		if err != nil {
			return err
		}
		if err := cred.RequireAuth(); err != nil {
			return err
		}
		client := newClient(cfg, cred)
		resp, err := client.ListDeployments(background(), args[0])
		if err != nil {
			return fmt.Errorf("list deployments: %w", err)
		}
		if len(resp.Deployments) == 0 {
			fmt.Println("No deployments found for this project.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tBRANCH\tSTATUS\tCOMMIT\tCREATED")
		fmt.Fprintln(w, "──\t──────\t──────\t──────\t───────")
		for _, d := range resp.Deployments {
			sha := d.CommitSHA
			if len(sha) > 7 {
				sha = sha[:7]
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				d.ID, d.Branch, d.Status, sha, d.CreatedAt.Format("Jan 2 15:04"))
		}
		_ = w.Flush()
		fmt.Printf("\nTotal: %d deployment(s)\n", resp.Total)
		return nil
	},
}

var deploymentsGetCmd = &cobra.Command{
	Use:   "get <deploymentId>",
	Short: "Show details for a deployment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, cred, err := loadAll()
		if err != nil {
			return err
		}
		if err := cred.RequireAuth(); err != nil {
			return err
		}
		client := newClient(cfg, cred)
		d, err := client.GetDeployment(background(), args[0])
		if err != nil {
			return fmt.Errorf("get deployment: %w", err)
		}
		if d == nil {
			return fmt.Errorf("deployment not found")
		}
		fmt.Printf("ID:        %s\n", d.ID)
		fmt.Printf("Project:   %s\n", d.ProjectID)
		fmt.Printf("Branch:    %s\n", d.Branch)
		fmt.Printf("Commit:    %s\n", d.CommitSHA)
		if d.CommitMessage != "" {
			fmt.Printf("Message:   %s\n", d.CommitMessage)
		}
		fmt.Printf("Status:    %s\n", d.Status)
		if d.DeployURL != "" {
			fmt.Printf("URL:       %s\n", d.DeployURL)
		}
		if d.ErrorMessage != nil {
			fmt.Printf("Error:     %s\n", *d.ErrorMessage)
		}
		fmt.Printf("Created:   %s\n", d.CreatedAt.Format("Jan 2, 2006 15:04 MST"))
		return nil
	},
}

var deploymentsCancelCmd = &cobra.Command{
	Use:   "cancel <deploymentId>",
	Short: "Cancel a pending or queued deployment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, cred, err := loadAll()
		if err != nil {
			return err
		}
		if err := cred.RequireAuth(); err != nil {
			return err
		}
		client := newClient(cfg, cred)
		if err := client.CancelDeployment(background(), args[0]); err != nil {
			return fmt.Errorf("cancel deployment: %w", err)
		}
		fmt.Printf("Deployment %s cancelled.\n", args[0])
		return nil
	},
}

func init() {
	deploymentsCmd.AddCommand(deploymentsListCmd)
	deploymentsCmd.AddCommand(deploymentsGetCmd)
	deploymentsCmd.AddCommand(deploymentsCancelCmd)
	rootCmd.AddCommand(deploymentsCmd)
}
