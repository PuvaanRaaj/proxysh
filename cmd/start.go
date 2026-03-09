package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/PuvaanRaaj/proxysh/cert"
	"github.com/PuvaanRaaj/proxysh/config"
	"github.com/PuvaanRaaj/proxysh/launchd"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the proxysh daemon and install certificates",
	Long: `Start sets up proxysh for the first time:
  1. Generates a local CA certificate
  2. Installs the CA into your system trust store (requires sudo once)
  3. Installs a LaunchAgent so the proxy starts automatically on login
  4. Starts the daemon`,
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

		// 2. Install CA into system trust (needs sudo)
		if !cert.IsCAInstalled(cfg.Cert.CADir) {
			fmt.Println("Installing CA into system trust store (this may ask for your password)...")
			if err := cert.InstallCA(cfg.Cert.CADir); err != nil {
				return fmt.Errorf("install CA: %w", err)
			}
			fmt.Println("  CA installed and trusted.")
		} else {
			fmt.Println("CA already trusted by system.")
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

		// 4. Install LaunchAgent
		binPath, err := os.Executable()
		if err != nil {
			binPath, _ = exec.LookPath("proxysh")
		}
		binPath, _ = filepath.Abs(binPath)

		plistPath := config.DefaultLaunchAgentPlist()
		if err := os.MkdirAll(filepath.Dir(plistPath), 0755); err != nil {
			return err
		}

		if err := launchd.Write(launchd.PlistData{
			BinPath:    binPath,
			ConfigPath: cfgPath,
			LogFile:    cfg.Daemon.LogFile,
		}, plistPath); err != nil {
			return fmt.Errorf("write plist: %w", err)
		}

		// 5. Load LaunchAgent
		if launchd.IsLoaded() {
			launchd.Unload(plistPath) //nolint:errcheck
		}
		if err := launchd.Load(plistPath); err != nil {
			return fmt.Errorf("load agent: %w", err)
		}

		fmt.Println("\nproxysh daemon started!")
		fmt.Printf("Proxy listening on https://127.0.0.1:%d\n", cfg.Daemon.ListenPort)
		fmt.Println("\nRun 'proxysh up <name> <port>' to add a .test domain.")
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
