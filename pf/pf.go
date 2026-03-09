package pf

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	anchorName  = "com.proxysh"
	anchorFile  = "/etc/pf.anchors/com.proxysh"
	daemonPlist = "/Library/LaunchDaemons/com.proxysh.pf.plist"
)

const daemonPlistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
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

// rule returns the pf rdr rule for the given port.
func rule(port int) string {
	return fmt.Sprintf("rdr pass on lo0 proto tcp from any to any port 443 -> 127.0.0.1 port %d", port)
}

// Enable installs the pf redirect rule for 443 → port immediately.
func Enable(port int) error {
	cmd := exec.Command("sudo", "pfctl", "-ef", "-")
	cmd.Stdin = strings.NewReader(rule(port) + "\n")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// InstallDaemon writes the anchor file and LaunchDaemon plist so the
// redirect rule survives reboots. Requires sudo once.
func InstallDaemon(port int) error {
	// Write anchor file
	anchorContent := rule(port) + "\n"
	writeAnchor := exec.Command("sudo", "tee", anchorFile)
	writeAnchor.Stdin = strings.NewReader(anchorContent)
	writeAnchor.Stdout = os.NewFile(0, os.DevNull)
	writeAnchor.Stderr = os.Stderr
	if err := writeAnchor.Run(); err != nil {
		return fmt.Errorf("write anchor file: %w", err)
	}

	// Write LaunchDaemon plist
	writePlist := exec.Command("sudo", "tee", daemonPlist)
	writePlist.Stdin = strings.NewReader(daemonPlistTemplate)
	writePlist.Stdout = os.NewFile(0, os.DevNull)
	writePlist.Stderr = os.Stderr
	if err := writePlist.Run(); err != nil {
		return fmt.Errorf("write plist: %w", err)
	}

	// Load the daemon
	load := exec.Command("sudo", "launchctl", "load", "-w", daemonPlist)
	load.Stdout = os.Stdout
	load.Stderr = os.Stderr
	return load.Run()
}

// IsDaemonInstalled returns true if the LaunchDaemon plist exists.
func IsDaemonInstalled() bool {
	_, err := os.Stat(daemonPlist)
	return err == nil
}

// IsEnabled returns true if the redirect rule is currently active.
func IsEnabled(port int) bool {
	out, err := exec.Command("pfctl", "-s", "nat").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), fmt.Sprintf("-> 127.0.0.1 port %d", port))
}
