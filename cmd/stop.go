package cmd

import (
	"fmt"

	"github.com/PuvaanRaaj/devtun/autostart"
	"github.com/PuvaanRaaj/devtun/config"
	"github.com/PuvaanRaaj/devtun/ipc"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the devtun daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := ipc.NewClient(config.IPCSocketPath)
		if err := client.Shutdown(); err != nil {
			fmt.Printf("Warning: %v\n", err)
		}

		if err := autostart.Uninstall(); err != nil {
			fmt.Printf("Warning: could not stop auto-start service: %v\n", err)
		}

		fmt.Println("devtun daemon stopped.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
