package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/PuvaanRaaj/proxysh/autostart"
	"github.com/PuvaanRaaj/proxysh/cert"
	"github.com/PuvaanRaaj/proxysh/config"
	"github.com/PuvaanRaaj/proxysh/hosts"
	"github.com/PuvaanRaaj/proxysh/ipc"
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

		// Auto-start service
		check("Auto-start service installed",
			autostart.IsInstalled(),
			"Run 'proxysh start' to install the auto-start service")

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

		// Port redirect check (macOS/Linux only)
		if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
			fmt.Println()
			fmt.Println("Network:")
			pfOk := checkPortRedirect(cfg.Daemon.ListenPort)
			check(
				fmt.Sprintf("Port 443 → %d redirect", cfg.Daemon.ListenPort),
				pfOk,
				"Run 'proxysh start' to set up the port redirect",
			)
		}

		fmt.Printf("\n%d checks passed, %d failed\n", pass, fail)
		if fail > 0 {
			fmt.Println("\nRun 'proxysh start' to fix most issues.")
		}
		return nil
	},
}

func checkPortRedirect(port int) bool {
	portStr := fmt.Sprintf("%d", port)
	switch runtime.GOOS {
	case "darwin":
		out, err := exec.Command("pfctl", "-s", "nat").Output()
		if err != nil {
			return false
		}
		return strings.Contains(string(out), portStr)
	case "linux":
		out, err := exec.Command("sudo", "iptables", "-t", "nat", "-L", "OUTPUT", "-n").Output()
		if err != nil {
			return false
		}
		return strings.Contains(string(out), portStr)
	default:
		return false
	}
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
