package launchd

import (
	"fmt"
	"os"
	"os/exec"
	"text/template"
)

const plistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
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

type PlistData struct {
	BinPath    string
	ConfigPath string
	LogFile    string
}

// Write generates the launchd plist and writes it to plistPath.
func Write(data PlistData, plistPath string) error {
	tmpl, err := template.New("plist").Parse(plistTemplate)
	if err != nil {
		return err
	}
	f, err := os.Create(plistPath)
	if err != nil {
		return fmt.Errorf("create plist: %w", err)
	}
	defer f.Close()
	return tmpl.Execute(f, data)
}

// Load loads (starts) the launchd agent.
func Load(plistPath string) error {
	cmd := exec.Command("launchctl", "load", "-w", plistPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("launchctl load: %w\n%s", err, out)
	}
	return nil
}

// Unload unloads (stops) the launchd agent.
func Unload(plistPath string) error {
	cmd := exec.Command("launchctl", "unload", "-w", plistPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("launchctl unload: %w\n%s", err, out)
	}
	return nil
}

// IsLoaded returns true if the launchd agent is currently loaded.
func IsLoaded() bool {
	cmd := exec.Command("launchctl", "list", "com.PuvaanRaaj.proxysh")
	return cmd.Run() == nil
}
