package hosts

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const hostsFile = "/etc/hosts"
const marker = "# devtun"

// writeHosts writes content to /etc/hosts, falling back to sudo tee if
// the file is not writable by the current user.
func writeHosts(content string) error {
	err := os.WriteFile(hostsFile, []byte(content), 0644)
	if err == nil {
		return nil
	}
	if !errors.Is(err, os.ErrPermission) {
		return err
	}
	// Fall back to sudo tee
	cmd := exec.Command("sudo", "tee", hostsFile)
	cmd.Stdin = strings.NewReader(content)
	cmd.Stdout = os.NewFile(0, os.DevNull)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// AddEntry adds a 127.0.0.1 <domain> entry to /etc/hosts.
// Only escalates to sudo if the file is not writable by the current user.
func AddEntry(domain string) error {
	if HasEntry(domain) {
		return nil
	}

	f, err := os.ReadFile(hostsFile)
	if err != nil {
		return err
	}

	content := strings.TrimRight(string(f), "\n") + "\n" +
		fmt.Sprintf("127.0.0.1 %s %s\n", domain, marker)

	return writeHosts(content)
}

// RemoveEntry removes the devtun-managed entry for domain from /etc/hosts.
func RemoveEntry(domain string) error {
	if !HasEntry(domain) {
		return nil
	}

	f, err := os.ReadFile(hostsFile)
	if err != nil {
		return err
	}

	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(string(f)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, domain) && strings.Contains(line, marker) {
			continue
		}
		lines = append(lines, line)
	}

	return writeHosts(strings.Join(lines, "\n") + "\n")
}

// HasEntry returns true if /etc/hosts already has an entry for domain managed by devtun.
func HasEntry(domain string) bool {
	f, err := os.ReadFile(hostsFile)
	if err != nil {
		return false
	}
	scanner := bufio.NewScanner(strings.NewReader(string(f)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, domain) && strings.Contains(line, marker) {
			return true
		}
	}
	return false
}

// ListEntries returns all domains managed by devtun in /etc/hosts.
func ListEntries() []string {
	f, err := os.ReadFile(hostsFile)
	if err != nil {
		return nil
	}
	var domains []string
	scanner := bufio.NewScanner(strings.NewReader(string(f)))
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, marker) {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			domains = append(domains, fields[1])
		}
	}
	return domains
}
