package cmd

import (
	"fmt"

	"github.com/PuvaanRaaj/devtun/config"
	"github.com/PuvaanRaaj/devtun/ipc"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all active .test domains",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := ipc.NewClient(config.IPCSocketPath)
		resp, err := client.Status()
		if err != nil {
			// Daemon not running — show config instead
			return listFromConfig()
		}

		if len(resp.Domains) == 0 {
			fmt.Println("No active domains. Run 'devtun up <name> <port>' to add one.")
			return nil
		}

		fmt.Printf("%-30s  %-30s  %s\n", "DOMAIN", "TARGET", "STATUS")
		fmt.Printf("%-30s  %-30s  %s\n", "------", "------", "------")
		for _, d := range resp.Domains {
			status := "active"
			if !d.Active {
				status = "inactive"
			}
			fmt.Printf("https://%-22s  %-30s  %s\n", d.Domain, d.Target, status)
		}
		return nil
	},
}

func listFromConfig() error {
	cfgPath := resolveConfig()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}
	if len(cfg.Domains) == 0 {
		fmt.Println("No domains configured. Run 'devtun start' then 'devtun up <name> <port>'.")
		return nil
	}
	fmt.Println("(daemon not running — showing config)")
	fmt.Printf("%-30s  %s\n", "DOMAIN", "TARGET")
	fmt.Printf("%-30s  %s\n", "------", "------")
	for _, d := range cfg.Domains {
		fmt.Printf("https://%-22s  %s\n", d.Domain, d.Target)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(listCmd)
}
