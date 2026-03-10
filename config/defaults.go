package config

import (
	"os"
	"path/filepath"
)

const (
	DaemonListenPort = 8443
	IPCSocketPath    = "/tmp/devtun.sock"
	ConfigFileName   = ".devtun.yaml"
)

func DefaultConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "devtun")
}

func DefaultCADir() string {
	return filepath.Join(DefaultConfigDir(), "ca")
}

func DefaultCertDir() string {
	return filepath.Join(DefaultConfigDir(), "certs")
}

func DefaultLogFile() string {
	return filepath.Join(DefaultConfigDir(), "devtun.log")
}

func DefaultPIDFile() string {
	return filepath.Join(DefaultConfigDir(), "devtun.pid")
}

func DefaultLaunchAgentPlist() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents", "com.PuvaanRaaj.devtun.plist")
}
