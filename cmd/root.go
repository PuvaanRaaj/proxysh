package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "devtun",
	Short: "Local HTTPS development domains & tunneling toolkit",
	Long: `devtun — Local HTTPS Development Toolkit

Transform localhost ports into trusted HTTPS .test domains.
Share your local servers publicly via secure tunnels.

Examples:
  devtun up example 3000         # https://example.test → localhost:3000
  devtun list                  # list active domains
  devtun share --port 3000     # get a public URL
  devtun doctor                # check system health`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func SetVersion(v string) {
	rootCmd.Version = v
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: .devtun.yaml)")
}
