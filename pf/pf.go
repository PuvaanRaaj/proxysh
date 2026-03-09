// Package pf manages port 443 → daemon redirect rules.
// macOS: pf with a LaunchDaemon for persistence
// Linux: iptables
// Windows: not yet supported
package pf

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Enable sets up the 443 → port redirect for the current session.
func Enable(port int) error {
	switch runtime.GOOS {
	case "darwin":
		return enablePF(port)
	case "linux":
		return enableIPTables(port)
	default:
		return fmt.Errorf("port redirect not supported on %s — access the proxy directly on port %d", runtime.GOOS, port)
	}
}

// InstallDaemon makes the redirect rule persist across reboots.
func InstallDaemon(port int) error {
	switch runtime.GOOS {
	case "darwin":
		return installPFDaemon(port)
	case "linux":
		return installIPTablesPersist(port)
	default:
		return fmt.Errorf("persistent redirect not supported on %s", runtime.GOOS)
	}
}

// IsDaemonInstalled returns true if a persistent redirect is set up.
func IsDaemonInstalled() bool {
	switch runtime.GOOS {
	case "darwin":
		_, err := os.Stat(daemonPlist)
		return err == nil
	case "linux":
		_, err := os.Stat(iptablesPersistFile)
		return err == nil
	default:
		return false
	}
}

// IsEnabled returns true if the redirect rule is currently active.
func IsEnabled(port int) bool {
	switch runtime.GOOS {
	case "darwin":
		out, err := exec.Command("pfctl", "-s", "nat").Output()
		if err != nil {
			return false
		}
		return strings.Contains(string(out), fmt.Sprintf("-> 127.0.0.1 port %d", port))
	case "linux":
		out, err := exec.Command("sudo", "iptables", "-t", "nat", "-L", "OUTPUT", "-n").Output()
		if err != nil {
			return false
		}
		return strings.Contains(string(out), fmt.Sprintf("redir ports %d", port))
	default:
		return false
	}
}

// ---- macOS pf implementation ----

const (
	anchorFile  = "/etc/pf.anchors/com.proxysh"
	daemonPlist = "/Library/LaunchDaemons/com.proxysh.pf.plist"
)

const pfDaemonPlist = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.proxysh.pf</string>
    <key>ProgramArguments</key>
    <array>
        <string>/sbin/pfctl</string>
        <string>-ef</string>
        <string>/etc/pf.anchors/com.proxysh</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
</dict>
</plist>
`

func pfRule(port int) string {
	return fmt.Sprintf("rdr pass on lo0 proto tcp from any to any port 443 -> 127.0.0.1 port %d", port)
}

func enablePF(port int) error {
	cmd := exec.Command("sudo", "pfctl", "-ef", "-")
	cmd.Stdin = strings.NewReader(pfRule(port) + "\n")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func installPFDaemon(port int) error {
	writeAnchor := exec.Command("sudo", "tee", anchorFile)
	writeAnchor.Stdin = strings.NewReader(pfRule(port) + "\n")
	writeAnchor.Stdout = os.NewFile(0, os.DevNull)
	writeAnchor.Stderr = os.Stderr
	if err := writeAnchor.Run(); err != nil {
		return fmt.Errorf("write anchor: %w", err)
	}

	writePlist := exec.Command("sudo", "tee", daemonPlist)
	writePlist.Stdin = strings.NewReader(pfDaemonPlist)
	writePlist.Stdout = os.NewFile(0, os.DevNull)
	writePlist.Stderr = os.Stderr
	if err := writePlist.Run(); err != nil {
		return fmt.Errorf("write plist: %w", err)
	}

	load := exec.Command("sudo", "launchctl", "load", "-w", daemonPlist)
	load.Stdout = os.Stdout
	load.Stderr = os.Stderr
	return load.Run()
}

// ---- Linux iptables implementation ----

const iptablesPersistFile = "/etc/proxysh-iptables.conf"

func enableIPTables(port int) error {
	// Redirect incoming on lo
	r1 := exec.Command("sudo", "iptables", "-t", "nat", "-A", "PREROUTING",
		"-i", "lo", "-p", "tcp", "--dport", "443",
		"-j", "REDIRECT", "--to-port", fmt.Sprintf("%d", port))
	r1.Stderr = os.Stderr
	if err := r1.Run(); err != nil {
		return fmt.Errorf("iptables PREROUTING: %w", err)
	}
	// Redirect locally-initiated connections
	r2 := exec.Command("sudo", "iptables", "-t", "nat", "-A", "OUTPUT",
		"-o", "lo", "-p", "tcp", "--dport", "443",
		"-j", "REDIRECT", "--to-port", fmt.Sprintf("%d", port))
	r2.Stderr = os.Stderr
	return r2.Run()
}

func installIPTablesPersist(port int) error {
	if err := enableIPTables(port); err != nil {
		return err
	}
	// Try to persist via iptables-save (works on Debian/Ubuntu/RHEL)
	out, err := exec.Command("sudo", "iptables-save").Output()
	if err != nil {
		return nil // rules are active but not persistent — best effort
	}
	write := exec.Command("sudo", "tee", iptablesPersistFile)
	write.Stdin = strings.NewReader(string(out))
	write.Stdout = os.NewFile(0, os.DevNull)
	write.Stderr = os.Stderr
	return write.Run()
}
