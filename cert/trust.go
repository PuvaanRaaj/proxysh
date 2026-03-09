package cert

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// InstallCA installs the CA into the system/user trust store for the current OS.
func InstallCA(caDir string) error {
	caPath := filepath.Join(caDir, "ca.crt")
	home, _ := os.UserHomeDir()

	switch runtime.GOOS {
	case "darwin":
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
		installNSS(caPath, home) // Firefox
		return nil

	case "linux":
		if err := installLinuxSystemStore(caPath); err != nil {
			return err
		}
		installNSS(caPath, home) // Firefox + Chrome
		return nil

	default:
		return fmt.Errorf("unsupported OS: %s — manually trust: %s", runtime.GOOS, caPath)
	}
}

// installLinuxSystemStore installs the CA into the system trust store,
// supporting Debian/Ubuntu, RHEL/Fedora/CentOS, and Arch.
func installLinuxSystemStore(caPath string) error {
	// Debian/Ubuntu
	if _, err := exec.LookPath("update-ca-certificates"); err == nil {
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
		return update.Run()
	}

	// RHEL / Fedora / CentOS
	if _, err := exec.LookPath("update-ca-trust"); err == nil {
		dest := "/etc/pki/ca-trust/source/anchors/proxysh-ca.crt"
		cp := exec.Command("sudo", "cp", caPath, dest)
		cp.Stdin = os.Stdin
		cp.Stderr = os.Stderr
		if err := cp.Run(); err != nil {
			return fmt.Errorf("copy CA: %w", err)
		}
		update := exec.Command("sudo", "update-ca-trust", "extract")
		update.Stdin = os.Stdin
		update.Stderr = os.Stderr
		return update.Run()
	}

	// Arch / Manjaro
	if _, err := exec.LookPath("trust"); err == nil {
		cmd := exec.Command("sudo", "trust", "anchor", "--store", caPath)
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	return fmt.Errorf("no supported certificate manager found — manually trust: %s", caPath)
}

// installNSS adds the CA to Firefox and Chrome NSS databases using certutil.
// Silently skips if certutil is not available.
func installNSS(caPath, home string) {
	certutil, err := findCertutil(home)
	if err != nil {
		return
	}
	for _, db := range findNSSDatabases(home) {
		exec.Command(certutil, "-A",
			"-n", "proxysh CA",
			"-t", "CT,,",
			"-i", caPath,
			"-d", "sql:"+db,
		).Run() //nolint:errcheck
	}
}

// findCertutil looks for certutil in Firefox bundle or system PATH.
func findCertutil(home string) (string, error) {
	candidates := []string{
		// macOS Firefox bundle
		"/Applications/Firefox.app/Contents/MacOS/certutil",
		// Linux Firefox bundle
		"/usr/lib/firefox/certutil",
		"/usr/lib64/firefox/certutil",
		filepath.Join(home, ".local/lib/firefox/certutil"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}
	// Fall back to system-installed (brew install nss / apt install libnss3-tools)
	return exec.LookPath("certutil")
}

// findNSSDatabases returns NSS database directories for Firefox and Chrome.
func findNSSDatabases(home string) []string {
	var dbs []string

	// Firefox profiles — macOS
	if dirs := globProfiles(filepath.Join(home, "Library", "Application Support", "Firefox", "Profiles")); len(dirs) > 0 {
		dbs = append(dbs, dirs...)
	}
	// Firefox profiles — Linux
	if dirs := globProfiles(filepath.Join(home, ".mozilla", "firefox")); len(dirs) > 0 {
		dbs = append(dbs, dirs...)
	}

	// Chrome — macOS (uses system keychain, but NSS db exists too)
	if dirs := globProfiles(filepath.Join(home, "Library", "Application Support", "Google", "Chrome")); len(dirs) > 0 {
		dbs = append(dbs, dirs...)
	}
	// Chrome — Linux
	if db := filepath.Join(home, ".pki", "nssdb"); dirExists(db) {
		dbs = append(dbs, db)
	}
	// Chromium — Linux
	if db := filepath.Join(home, ".pki", "nssdb"); dirExists(db) {
		dbs = append(dbs, db)
	}

	return dbs
}

func globProfiles(profilesDir string) []string {
	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return nil
	}
	var dirs []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		n := e.Name()
		if strings.HasSuffix(n, ".default-release") ||
			strings.HasSuffix(n, ".default") ||
			strings.Contains(n, "release") ||
			strings.Contains(n, "Profile") {
			dirs = append(dirs, filepath.Join(profilesDir, n))
		}
	}
	return dirs
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// UninstallCA removes the CA from the trust store.
func UninstallCA(caDir string) error {
	caPath := filepath.Join(caDir, "ca.crt")
	home, _ := os.UserHomeDir()

	switch runtime.GOOS {
	case "darwin":
		cmd := exec.Command("security", "remove-trusted-cert", caPath)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to remove CA: %w\n%s", err, out)
		}
		removeNSS(home)
		return nil

	case "linux":
		// Debian/Ubuntu
		if _, err := exec.LookPath("update-ca-certificates"); err == nil {
			exec.Command("sudo", "rm", "-f", "/usr/local/share/ca-certificates/proxysh-ca.crt").Run() //nolint:errcheck
			exec.Command("sudo", "update-ca-certificates", "--fresh").Run()                            //nolint:errcheck
		}
		// RHEL/Fedora
		if _, err := exec.LookPath("update-ca-trust"); err == nil {
			exec.Command("sudo", "rm", "-f", "/etc/pki/ca-trust/source/anchors/proxysh-ca.crt").Run() //nolint:errcheck
			exec.Command("sudo", "update-ca-trust", "extract").Run()                                   //nolint:errcheck
		}
		removeNSS(home)
		return nil

	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func removeNSS(home string) {
	certutil, err := findCertutil(home)
	if err != nil {
		return
	}
	for _, db := range findNSSDatabases(home) {
		exec.Command(certutil, "-D", "-n", "proxysh CA", "-d", "sql:"+db).Run() //nolint:errcheck
	}
}

// IsCAInstalled checks if the CA is currently trusted.
func IsCAInstalled(caDir string) bool {
	caPath := filepath.Join(caDir, "ca.crt")
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("security", "verify-cert", "-c", caPath).Run() == nil
	case "linux":
		// Check if the cert file exists in any known system store
		candidates := []string{
			"/usr/local/share/ca-certificates/proxysh-ca.crt",
			"/etc/pki/ca-trust/source/anchors/proxysh-ca.crt",
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				return true
			}
		}
		return false
	default:
		return false
	}
}
