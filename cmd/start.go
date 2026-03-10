package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/PuvaanRaaj/devtun/autostart"
	"github.com/PuvaanRaaj/devtun/cert"
	"github.com/PuvaanRaaj/devtun/config"
	"github.com/PuvaanRaaj/devtun/pf"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the devtun daemon and install certificates",
	Long: `Start sets up devtun for the first time:
  1. Generates a local CA certificate
  2. Installs the CA into your trust store (no sudo on macOS)
  3. Installs a service so the proxy starts automatically on login
  4. Sets up port 443 redirect (requires sudo once)
  5. Starts the daemon`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath := resolveConfig()
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return err
		}

		// 1. Generate CA if needed
		if !cert.CAExists(cfg.Cert.CADir) {
			fmt.Println("Generating local CA certificate...")
			ca, err := cert.GenerateCA(cfg.Cert.CACommonName)
			if err != nil {
				return fmt.Errorf("generate CA: %w", err)
			}
			if err := cert.SaveCA(ca, cfg.Cert.CADir); err != nil {
				return fmt.Errorf("save CA: %w", err)
			}
			fmt.Printf("  CA saved to %s\n", cfg.Cert.CADir)
		} else {
			fmt.Println("CA certificate already exists, skipping.")
		}

		// 2. Install CA into trust store
		if !cert.IsCAInstalled(cfg.Cert.CADir) {
			fmt.Println("Installing CA into trust store...")
			if err := cert.InstallCA(cfg.Cert.CADir); err != nil {
				return fmt.Errorf("install CA: %w", err)
			}
			fmt.Println("  CA installed and trusted.")
		} else {
			fmt.Println("CA already trusted.")
		}

		// 3. Generate domain certs for any already-configured domains
		if len(cfg.Domains) > 0 {
			ca, err := cert.LoadCA(cfg.Cert.CADir)
			if err != nil {
				return err
			}
			for _, d := range cfg.Domains {
				if err := cert.EnsureDomainCert(ca, cfg.Cert.CertDir, d.Domain, cfg.Cert.DaysValid); err != nil {
					fmt.Printf("  Warning: could not generate cert for %s: %v\n", d.Domain, err)
				}
			}
		}

		// 4. Install auto-start service
		binPath, err := os.Executable()
		if err != nil {
			binPath, _ = exec.LookPath("devtun")
		}
		binPath, _ = filepath.Abs(binPath)

		if err := autostart.Install(autostart.Config{
			BinPath:    binPath,
			ConfigPath: cfgPath,
			LogFile:    cfg.Daemon.LogFile,
		}); err != nil {
			fmt.Printf("  Warning: could not install auto-start service: %v\n", err)
			fmt.Println("  Start the daemon manually with: devtun daemon")
		}

		// 5. Set up port 443 redirect (macOS + Linux only, requires sudo once)
		if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
			if !pf.IsDaemonInstalled() {
				fmt.Println("Setting up port 443 redirect (requires sudo, persists across reboots)...")
				if err := pf.InstallDaemon(cfg.Daemon.ListenPort); err != nil {
					fmt.Printf("  Warning: could not set up port redirect: %v\n", err)
					fmt.Printf("  You can still access domains on port 8443 directly.\n")
				} else {
					fmt.Println("  Port redirect installed and active.")
				}
			} else if !pf.IsEnabled(cfg.Daemon.ListenPort) {
				pf.Enable(cfg.Daemon.ListenPort) //nolint:errcheck
				fmt.Println("Port redirect active.")
			} else {
				fmt.Println("Port redirect already active.")
			}
		}

		fmt.Println("\ndevtun daemon started!")
		fmt.Printf("Proxy listening on https://127.0.0.1:%d\n", cfg.Daemon.ListenPort)
		fmt.Println("\nRun 'devtun up <name> <port>' to add a .test domain.")
		return nil
	},
}

func resolveConfig() string {
	if cfgFile != "" {
		return cfgFile
	}
	return config.FindConfigFile()
}

func init() {
	rootCmd.AddCommand(startCmd)
}
