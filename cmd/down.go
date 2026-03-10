package cmd

import (
	"fmt"

	"github.com/PuvaanRaaj/devtun/cert"
	"github.com/PuvaanRaaj/devtun/config"
	"github.com/PuvaanRaaj/devtun/hosts"
	"github.com/PuvaanRaaj/devtun/ipc"
	"github.com/spf13/cobra"
)

var downCmd = &cobra.Command{
	Use:   "down <name>",
	Short: "Remove a .test domain",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		domain := name + ".test"
		cfgPath := resolveConfig()

		cfg, err := config.Load(cfgPath)
		if err != nil {
			return err
		}

		if !cfg.RemoveDomain(domain) {
			return fmt.Errorf("domain %s not found in config", domain)
		}

		if err := cfg.Save(cfgPath); err != nil {
			return err
		}

		// Remove cert
		cert.RemoveDomainCert(cfg.Cert.CertDir, domain) //nolint:errcheck

		// Remove hosts entry
		fmt.Printf("Removing %s from /etc/hosts...\n", domain)
		if err := hosts.RemoveEntry(domain); err != nil {
			fmt.Printf("Warning: could not remove hosts entry: %v\n", err)
		}

		// Reload daemon
		client := ipc.NewClient(config.IPCSocketPath)
		if err := client.Reload(); err != nil {
			fmt.Printf("Warning: could not reload daemon: %v\n", err)
		}

		fmt.Printf("Removed %s\n", domain)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(downCmd)
}
