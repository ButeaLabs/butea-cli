package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check the Butea API health status",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, cred, err := loadAll()
		if err != nil {
			return err
		}
		client := newClient(cfg, cred)
		h, err := client.Health(background())
		if err != nil {
			fmt.Printf("API unreachable at %s\nError: %v\n", cfg.APIURL, err)
			return nil
		}
		icon := "✓"
		if h.Status != "healthy" {
			icon = "⚠"
		}
		fmt.Printf("%s API status: %s (v%s)\n\n", icon, h.Status, h.Version)
		fmt.Printf("Endpoint: %s\n", cfg.APIURL)
		fmt.Printf("Checked:  %s\n\n", h.Timestamp.Format("Jan 2, 2006 15:04:05 MST"))
		if len(h.Services) > 0 {
			fmt.Println("Services:")
			for name, svc := range h.Services {
				svcIcon := "✓"
				if svc.Status != "healthy" {
					svcIcon = "✗"
				}
				detail := svc.Latency
				if svc.Error != "" {
					detail = svc.Error
				}
				fmt.Printf("  %s %s: %s", svcIcon, name, svc.Status)
				if detail != "" {
					fmt.Printf(" (%s)", detail)
				}
				fmt.Println()
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(healthCmd)
}
