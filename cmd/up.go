package cmd

import (
	"fmt"
	"strconv"

	"github.com/PuvaanRaaj/devtun/cert"
	"github.com/PuvaanRaaj/devtun/config"
	"github.com/PuvaanRaaj/devtun/hosts"
	"github.com/PuvaanRaaj/devtun/ipc"
	"github.com/spf13/cobra"
)

var upCmd = &cobra.Command{
	Use:   "up <name> <port>",
	Short: "Add a .test domain mapped to a local port",
	Long: `Add a domain mapping and immediately make it available over HTTPS.

Examples:
  devtun up example 3000        # creates https://example.test → localhost:3000
  devtun up api 8080          # creates https://api.test → localhost:8080`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		port := args[1]

		if _, err := strconv.Atoi(port); err != nil {
			return fmt.Errorf("port must be a number, got: %s", port)
		}

		domain := name + ".test"
		target := "http://localhost:" + port
		cfgPath := resolveConfig()

		cfg, err := config.Load(cfgPath)
		if err != nil {
			return err
		}

		// Load or generate CA
		if !cert.CAExists(cfg.Cert.CADir) {
			return fmt.Errorf("devtun is not set up — run 'devtun start' first")
		}
		ca, err := cert.LoadCA(cfg.Cert.CADir)
		if err != nil {
			return fmt.Errorf("load CA: %w", err)
		}

		// Generate domain cert
		fmt.Printf("Generating certificate for %s...\n", domain)
		if err := cert.EnsureDomainCert(ca, cfg.Cert.CertDir, domain, cfg.Cert.DaysValid); err != nil {
			return fmt.Errorf("generate cert: %w", err)
		}

		// Update /etc/hosts
		fmt.Printf("Adding %s to /etc/hosts (may ask for your password)...\n", domain)
		if err := hosts.AddEntry(domain); err != nil {
			return fmt.Errorf("hosts: %w", err)
		}

		// Update config
		cfg.AddDomain(domain, target)
		if err := cfg.Save(cfgPath); err != nil {
			return fmt.Errorf("save config: %w", err)
		}

		// Reload daemon
		client := ipc.NewClient(config.IPCSocketPath)
		if err := client.Reload(); err != nil {
			fmt.Printf("Warning: could not reload daemon: %v\n", err)
			fmt.Println("  Run 'devtun start' if the daemon is not running.")
		}

		fmt.Printf("\nhttps://%s → localhost:%s\n", domain, port)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(upCmd)
}
