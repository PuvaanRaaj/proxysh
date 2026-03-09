package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/PuvaanRaaj/proxysh/config"
	"github.com/spf13/cobra"
)

var (
	logsFollow bool
	logsLines  int
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View daemon logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath := resolveConfig()
		cfg, err := config.Load(cfgPath)
		if err != nil {
			return err
		}

		logFile := cfg.Daemon.LogFile
		if _, err := os.Stat(logFile); os.IsNotExist(err) {
			return fmt.Errorf("log file not found: %s\nIs the daemon running? Try 'proxysh start'", logFile)
		}

		tailArgs := []string{fmt.Sprintf("-n%d", logsLines)}
		if logsFollow {
			tailArgs = append(tailArgs, "-f")
		}
		tailArgs = append(tailArgs, logFile)

		c := exec.Command("tail", tailArgs...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		return c.Run()
	},
}

func init() {
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Follow log output")
	logsCmd.Flags().IntVarP(&logsLines, "lines", "n", 50, "Number of lines to show")
	rootCmd.AddCommand(logsCmd)
}
