package cmd

import (
	"fmt"

	"github.com/PuvaanRaaj/proxysh/autostart"
	"github.com/PuvaanRaaj/proxysh/config"
	"github.com/PuvaanRaaj/proxysh/ipc"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the proxysh daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := ipc.NewClient(config.IPCSocketPath)
		if err := client.Shutdown(); err != nil {
			fmt.Printf("Warning: %v\n", err)
		}

		if err := autostart.Uninstall(); err != nil {
			fmt.Printf("Warning: could not stop auto-start service: %v\n", err)
		}

		fmt.Println("proxysh daemon stopped.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}
