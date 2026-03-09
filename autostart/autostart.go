// Package autostart manages daemon auto-start across platforms.
// macOS: launchd LaunchAgent
// Linux: systemd user service
// Windows: not yet supported
package autostart

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"text/template"
)

// ---- macOS launchd ----

const launchdLabel = "com.PuvaanRaaj.proxysh"

const launchdTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.PuvaanRaaj.proxysh</string>
    <key>ProgramArguments</key>
    <array>
        <string>{{.BinPath}}</string>
        <string>daemon</string>
        <string>--config</string>
        <string>{{.ConfigPath}}</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>{{.LogFile}}</string>
    <key>StandardErrorPath</key>
    <string>{{.LogFile}}</string>
</dict>
</plist>
`

// ---- Linux systemd ----

const systemdTemplate = `[Unit]
Description=proxysh local HTTPS proxy daemon
After=network.target

[Service]
ExecStart={{.BinPath}} daemon --config {{.ConfigPath}}
Restart=on-failure
StandardOutput=append:{{.LogFile}}
StandardError=append:{{.LogFile}}

[Install]
WantedBy=default.target
`

// Config holds the data needed to generate service files.
type Config struct {
	BinPath    string
	ConfigPath string
	LogFile    string
}

// Install writes and loads the appropriate service for the current OS.
func Install(cfg Config) error {
	switch runtime.GOOS {
	case "darwin":
		return installLaunchd(cfg)
	case "linux":
		return installSystemd(cfg)
	default:
		return fmt.Errorf("auto-start not supported on %s — start the daemon manually with: proxysh daemon", runtime.GOOS)
	}
}

// Uninstall removes and unloads the service.
func Uninstall() error {
	switch runtime.GOOS {
	case "darwin":
		return uninstallLaunchd()
	case "linux":
		return uninstallSystemd()
	default:
		return nil
	}
}

// IsInstalled returns true if the service is currently loaded/active.
func IsInstalled() bool {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("launchctl", "list", launchdLabel).Run() == nil
	case "linux":
		return exec.Command("systemctl", "--user", "is-enabled", "--quiet", "proxysh").Run() == nil
	default:
		return false
	}
}

// ---- macOS implementation ----

func launchdPlistPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents", launchdLabel+".plist")
}

func installLaunchd(cfg Config) error {
	plistPath := launchdPlistPath()
	if err := os.MkdirAll(filepath.Dir(plistPath), 0755); err != nil {
		return err
	}
	if err := writeTemplate(plistPath, launchdTemplate, cfg); err != nil {
		return err
	}
	// Unload first if already loaded
	exec.Command("launchctl", "unload", "-w", plistPath).Run() //nolint:errcheck
	out, err := exec.Command("launchctl", "load", "-w", plistPath).CombinedOutput()
	if err != nil {
		return fmt.Errorf("launchctl load: %w\n%s", err, out)
	}
	return nil
}

func uninstallLaunchd() error {
	plistPath := launchdPlistPath()
	exec.Command("launchctl", "unload", "-w", plistPath).Run() //nolint:errcheck
	return os.Remove(plistPath)
}

// ---- Linux implementation ----

func systemdServicePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "systemd", "user", "proxysh.service")
}

func installSystemd(cfg Config) error {
	servicePath := systemdServicePath()
	if err := os.MkdirAll(filepath.Dir(servicePath), 0755); err != nil {
		return err
	}
	if err := writeTemplate(servicePath, systemdTemplate, cfg); err != nil {
		return err
	}
	exec.Command("systemctl", "--user", "daemon-reload").Run()          //nolint:errcheck
	exec.Command("systemctl", "--user", "enable", "proxysh").Run()      //nolint:errcheck
	out, err := exec.Command("systemctl", "--user", "start", "proxysh").CombinedOutput()
	if err != nil {
		return fmt.Errorf("systemctl start: %w\n%s", err, out)
	}
	return nil
}

func uninstallSystemd() error {
	exec.Command("systemctl", "--user", "stop", "proxysh").Run()    //nolint:errcheck
	exec.Command("systemctl", "--user", "disable", "proxysh").Run() //nolint:errcheck
	return os.Remove(systemdServicePath())
}

// ---- helpers ----

func writeTemplate(path, tmplStr string, data Config) error {
	tmpl, err := template.New("").Parse(tmplStr)
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return tmpl.Execute(f, data)
}
