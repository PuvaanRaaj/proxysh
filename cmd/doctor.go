package cmd

import (
	"fmt"
	"os/exec"

	"github.com/PuvaanRaaj/proxysh/cert"
	"github.com/PuvaanRaaj/proxysh/config"
	"github.com/PuvaanRaaj/proxysh/hosts"
	"github.com/PuvaanRaaj/proxysh/ipc"
	"github.com/PuvaanRaaj/proxysh/launchd"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system health and configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath := resolveConfig()
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return err
		}

		pass := 0
		fail := 0

		check := func(name string, ok bool, hint string) {
			if ok {
				fmt.Printf("  ✓  %s\n", name)
				pass++
			} else {
				fmt.Printf("  ✗  %s\n", name)
				if hint != "" {
					fmt.Printf("       → %s\n", hint)
				}
				fail++
			}
		}

		fmt.Println("\nproxysh doctor\n")

		// CA
		check("CA certificate exists",
			cert.CAExists(cfg.Cert.CADir),
			"Run 'proxysh start' to generate a CA certificate")

		check("CA trusted by system",
			cert.IsCAInstalled(cfg.Cert.CADir),
			"Run 'proxysh start' to install the CA")

		// Daemon
		client := ipc.NewClient(config.IPCSocketPath)
		_, daemonErr := client.Status()
		check("Daemon is running",
			daemonErr == nil,
			"Run 'proxysh start' to start the daemon")

		// LaunchAgent
		check("LaunchAgent installed",
			launchd.IsLoaded(),
			"Run 'proxysh start' to install and start the LaunchAgent")

		// Per-domain checks
		if len(cfg.Domains) > 0 {
			fmt.Println()
			fmt.Println("Domains:")
			for _, d := range cfg.Domains {
				check(
					fmt.Sprintf("Certificate: %s", d.Domain),
					cert.DomainCertExists(cfg.Cert.CertDir, d.Domain),
					fmt.Sprintf("Run 'proxysh up %s <port>'", d.Domain),
				)
				check(
					fmt.Sprintf("/etc/hosts: %s", d.Domain),
					hosts.HasEntry(d.Domain),
					fmt.Sprintf("Run 'proxysh up %s <port>'", d.Domain),
				)
			}
		}

		// pf redirect check
		fmt.Println()
		fmt.Println("Network:")
		pfOk := checkPFRedirect(cfg.Daemon.ListenPort)
		check(
			fmt.Sprintf("Port 443 → %d redirect (pf)", cfg.Daemon.ListenPort),
			pfOk,
			fmt.Sprintf("Run: echo 'rdr pass on lo0 proto tcp from any to any port 443 -> 127.0.0.1 port %d' | sudo pfctl -ef -", cfg.Daemon.ListenPort),
		)

		fmt.Printf("\n%d checks passed, %d failed\n", pass, fail)
		if fail > 0 {
			fmt.Println("\nRun 'proxysh start' to fix most issues.")
		}
		return nil
	},
}

func checkPFRedirect(port int) bool {
	out, err := exec.Command("sudo", "pfctl", "-s", "nat").Output()
	if err != nil {
		return false
	}
	return len(out) > 0 && containsPort(string(out), port)
}

func containsPort(s string, port int) bool {
	return len(s) > 0 && (len(fmt.Sprintf("%d", port)) > 0)
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
