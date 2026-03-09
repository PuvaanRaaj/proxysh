package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "proxysh",
	Short: "Local HTTPS development domains & tunneling toolkit",
	Long: `proxysh — Local HTTPS Development Toolkit

Transform localhost ports into trusted HTTPS .test domains.
Share your local servers publicly via secure tunnels.

Examples:
  proxysh up myapp 3000         # https://myapp.test → localhost:3000
  proxysh list                  # list active domains
  proxysh share --port 3000     # get a public URL
  proxysh doctor                # check system health`,
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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: .proxysh.yaml)")
}
