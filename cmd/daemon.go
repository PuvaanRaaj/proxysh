package cmd

import (
	"fmt"

	"github.com/PuvaanRaaj/proxysh/config"
	"github.com/PuvaanRaaj/proxysh/daemon"
	"github.com/spf13/cobra"
)

// daemonCmd is the hidden internal command that runs the actual proxy server.
// It is invoked by the LaunchAgent plist, not directly by users.
var daemonCmd = &cobra.Command{
	Use:    "daemon",
	Short:  "Run the proxy daemon (internal use)",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath := resolveConfig()
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		return daemon.Run(cfg, cfgPath)
	},
}

func init() {
	rootCmd.AddCommand(daemonCmd)
}
