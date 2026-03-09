package cert

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// InstallCA installs the CA certificate into the user login keychain.
// No sudo required — the login keychain is trusted by Safari, Chrome, and curl.
// Also installs into Firefox's NSS store if Firefox is present.
func InstallCA(caDir string) error {
	caPath := filepath.Join(caDir, "ca.crt")
	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("find home dir: %w", err)
		}
		loginKeychain := filepath.Join(home, "Library", "Keychains", "login.keychain-db")
		cmd := exec.Command("security", "add-trusted-cert",
			"-r", "trustRoot",
			"-k", loginKeychain,
			caPath,
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to install CA: %w\n%s", err, out)
		}
		// Also trust in Firefox if installed
		installFirefox(caPath, home)
		return nil
	case "linux":
		// Try update-ca-certificates (Debian/Ubuntu)
		dest := "/usr/local/share/ca-certificates/proxysh-ca.crt"
		cp := exec.Command("sudo", "cp", caPath, dest)
		cp.Stdin = os.Stdin
		cp.Stderr = os.Stderr
		if err := cp.Run(); err != nil {
			return fmt.Errorf("copy CA: %w", err)
		}
		update := exec.Command("sudo", "update-ca-certificates")
		update.Stdin = os.Stdin
		update.Stderr = os.Stderr
		if err := update.Run(); err != nil {
			return fmt.Errorf("update-ca-certificates: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported OS: %s — please manually trust %s", runtime.GOOS, caPath)
	}
}

// installFirefox adds the CA to all Firefox profile NSS databases using certutil.
// Silently skips if Firefox is not installed or certutil is not available.
func installFirefox(caPath, home string) {
	certutil, err := findCertutil()
	if err != nil {
		return
	}

	profiles := findFirefoxProfiles(home)
	for _, profile := range profiles {
		exec.Command(certutil, "-A",
			"-n", "proxysh CA",
			"-t", "CT,,",
			"-i", caPath,
			"-d", "sql:"+profile,
		).Run() //nolint:errcheck
	}
}

// findCertutil looks for certutil in common locations.
func findCertutil() (string, error) {
	// Check bundled with Firefox first
	bundled := "/Applications/Firefox.app/Contents/MacOS/certutil"
	if _, err := os.Stat(bundled); err == nil {
		return bundled, nil
	}
	// Fall back to system-installed (brew install nss)
	return exec.LookPath("certutil")
}

// findFirefoxProfiles returns paths to all Firefox profile directories.
func findFirefoxProfiles(home string) []string {
	profilesDir := filepath.Join(home, "Library", "Application Support", "Firefox", "Profiles")
	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return nil
	}
	var profiles []string
	for _, e := range entries {
		if e.IsDir() && (strings.HasSuffix(e.Name(), ".default-release") ||
			strings.HasSuffix(e.Name(), ".default") ||
			strings.Contains(e.Name(), "release")) {
			profiles = append(profiles, filepath.Join(profilesDir, e.Name()))
		}
	}
	return profiles
}

// UninstallCA removes the CA certificate from the system trust store.
func UninstallCA(caDir string) error {
	caPath := filepath.Join(caDir, "ca.crt")
	switch runtime.GOOS {
	case "darwin":
		cmd := exec.Command("security", "remove-trusted-cert", caPath)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to remove CA: %w\n%s", err, out)
		}
		// Also remove from Firefox
		home, _ := os.UserHomeDir()
		certutil, err := findCertutil()
		if err == nil {
			for _, profile := range findFirefoxProfiles(home) {
				exec.Command(certutil, "-D", "-n", "proxysh CA", "-d", "sql:"+profile).Run() //nolint:errcheck
			}
		}
		return nil
	case "linux":
		rm := exec.Command("sudo", "rm", "-f", "/usr/local/share/ca-certificates/proxysh-ca.crt")
		rm.Run()
		update := exec.Command("sudo", "update-ca-certificates", "--fresh")
		update.Run()
		return nil
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

// IsCAInstalled checks if the CA is currently trusted by the system.
func IsCAInstalled(caDir string) bool {
	caPath := filepath.Join(caDir, "ca.crt")
	switch runtime.GOOS {
	case "darwin":
		cmd := exec.Command("security", "verify-cert", "-c", caPath)
		return cmd.Run() == nil
	case "linux":
		_, err := exec.LookPath("update-ca-certificates")
		return err == nil
	default:
		return false
	}
}
