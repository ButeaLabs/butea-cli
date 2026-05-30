package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Manage your Butea projects",
}

var projectsListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, cred, err := loadAll()
		if err != nil {
			return err
		}
		if err := cred.RequireAuth(); err != nil {
			return err
		}
		client := newClient(cfg, cred)
		resp, err := client.ListProjects(background())
		if err != nil {
			return fmt.Errorf("list projects: %w", err)
		}
		if len(resp.Projects) == 0 {
			fmt.Printf("No projects found. Import one at %s/import\n", cfg.AppURL)
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tPROVIDER\tREPO\tBRANCH\tSTATUS")
		fmt.Fprintln(w, "──\t────\t────────\t────\t──────\t──────")
		for _, p := range resp.Projects {
			vis := "public"
			if p.IsPrivate {
				vis = "private"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s (%s)\t%s\t%s\n",
				p.ID, p.Name, p.Provider, p.RepoFullName, vis, p.DefaultBranch, p.Status)
		}
		_ = w.Flush()
		fmt.Printf("\nTotal: %d project(s)\n", resp.Total)
		return nil
	},
}

var projectsGetCmd = &cobra.Command{
	Use:   "get <projectId>",
	Short: "Show details for a project",
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
		p, err := client.GetProject(background(), args[0])
		if err != nil {
			return fmt.Errorf("get project: %w", err)
		}
		if p == nil {
			return fmt.Errorf("project not found")
		}
		desc := "<none>"
		if p.Description != nil {
			desc = *p.Description
		}
		fw := "<auto-detect>"
		if p.Framework != nil {
			fw = *p.Framework
		}
		fmt.Printf("ID:          %s\n", p.ID)
		fmt.Printf("Name:        %s\n", p.Name)
		fmt.Printf("Description: %s\n", desc)
		fmt.Printf("Provider:    %s\n", p.Provider)
		fmt.Printf("Repository:  %s\n", p.RepoFullName)
		fmt.Printf("Branch:      %s\n", p.DefaultBranch)
		fmt.Printf("Private:     %v\n", p.IsPrivate)
		fmt.Printf("Framework:   %s\n", fw)
		fmt.Printf("Status:      %s\n", p.Status)
		fmt.Printf("Created:     %s\n", p.CreatedAt.Format("Jan 2, 2006 15:04 MST"))
		return nil
	},
}

var projectsDeleteCmd = &cobra.Command{
	Use:     "delete <projectId>",
	Aliases: []string{"rm"},
	Short:   "Delete a project",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, cred, err := loadAll()
		if err != nil {
			return err
		}
		if err := cred.RequireAuth(); err != nil {
			return err
		}
		yes, _ := cmd.Flags().GetBool("yes")
		if !yes {
			answer, promptErr := prompt(fmt.Sprintf("Delete project %s? This cannot be undone [y/N]: ", args[0]))
			if promptErr != nil {
				return promptErr
			}
			if answer != "y" && answer != "Y" && answer != "yes" {
				fmt.Println("Aborted.")
				return nil
			}
		}
		client := newClient(cfg, cred)
		if err := client.DeleteProject(background(), args[0]); err != nil {
			return fmt.Errorf("delete project: %w", err)
		}
		fmt.Printf("Project %s deleted.\n", args[0])
		return nil
	},
}

func init() {
	projectsDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	projectsCmd.AddCommand(projectsListCmd)
	projectsCmd.AddCommand(projectsGetCmd)
	projectsCmd.AddCommand(projectsDeleteCmd)
	rootCmd.AddCommand(projectsCmd)
}
