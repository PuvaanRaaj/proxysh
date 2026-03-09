package config

import (
	"os"
	"path/filepath"
)

const (
	DaemonListenPort = 8443
	IPCSocketPath    = "/tmp/proxysh.sock"
	ConfigFileName   = ".proxysh.yaml"
)

func DefaultConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "proxysh")
}

func DefaultCADir() string {
	return filepath.Join(DefaultConfigDir(), "ca")
}

func DefaultCertDir() string {
	return filepath.Join(DefaultConfigDir(), "certs")
}

func DefaultLogFile() string {
	return filepath.Join(DefaultConfigDir(), "proxysh.log")
}

func DefaultPIDFile() string {
	return filepath.Join(DefaultConfigDir(), "proxysh.pid")
}

func DefaultLaunchAgentPlist() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents", "com.PuvaanRaaj.proxysh.plist")
}
