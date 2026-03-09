package cert

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
)

// InstallCA installs the CA certificate into the system trust store.
// On macOS this requires one sudo prompt.
func InstallCA(caDir string) error {
	caPath := filepath.Join(caDir, "ca.crt")
	switch runtime.GOOS {
	case "darwin":
		cmd := exec.Command("sudo", "security", "add-trusted-cert",
			"-d",
			"-r", "trustRoot",
			"-k", "/Library/Keychains/System.keychain",
			caPath,
		)
		cmd.Stdin = nil
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to install CA: %w\n%s", err, out)
		}
		return nil
	case "linux":
		// Try update-ca-certificates (Debian/Ubuntu)
		dest := "/usr/local/share/ca-certificates/proxysh-ca.crt"
		cp := exec.Command("sudo", "cp", caPath, dest)
		if out, err := cp.CombinedOutput(); err != nil {
			return fmt.Errorf("copy CA: %w\n%s", err, out)
		}
		update := exec.Command("sudo", "update-ca-certificates")
		if out, err := update.CombinedOutput(); err != nil {
			return fmt.Errorf("update-ca-certificates: %w\n%s", err, out)
		}
		return nil
	default:
		return fmt.Errorf("unsupported OS: %s — please manually trust %s", runtime.GOOS, caPath)
	}
}

// UninstallCA removes the CA certificate from the system trust store.
func UninstallCA(caDir string) error {
	caPath := filepath.Join(caDir, "ca.crt")
	switch runtime.GOOS {
	case "darwin":
		cmd := exec.Command("sudo", "security", "remove-trusted-cert", "-d", caPath)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to remove CA: %w\n%s", err, out)
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
