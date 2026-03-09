package hosts

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const hostsFile = "/etc/hosts"
const marker = "# proxysh"

// AddEntry adds a 127.0.0.1 <domain> entry to /etc/hosts using sudo.
func AddEntry(domain string) error {
	if HasEntry(domain) {
		return nil
	}
	entry := fmt.Sprintf("127.0.0.1 %s %s", domain, marker)
	cmd := exec.Command("sudo", "sh", "-c",
		fmt.Sprintf("echo %q >> %s", entry, hostsFile))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add hosts entry: %w\n%s", err, out)
	}
	return nil
}

// RemoveEntry removes the proxysh-managed entry for domain from /etc/hosts using sudo.
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

	newContent := strings.Join(lines, "\n") + "\n"
	cmd := exec.Command("sudo", "sh", "-c",
		fmt.Sprintf("cat > %s", hostsFile))
	cmd.Stdin = strings.NewReader(newContent)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to update hosts: %w\n%s", err, out)
	}
	return nil
}

// HasEntry returns true if /etc/hosts already has an entry for domain managed by proxysh.
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

// ListEntries returns all domains managed by proxysh in /etc/hosts.
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
