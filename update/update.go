package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const releaseAPI = "https://api.github.com/repos/PuvaanRaaj/proxysh/releases/latest"

// Check fetches the latest release from GitHub and prints a notice if a newer
// version is available. Runs synchronously — call it in a goroutine.
func Check(current string) {
	if current == "dev" || current == "" {
		return
	}

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(releaseAPI)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	curr := strings.TrimPrefix(current, "v")

	if latest != "" && latest != curr {
		fmt.Printf("\n  A new version of proxysh is available: v%s → v%s\n", curr, latest)
		fmt.Println("  Update: curl -sL https://proxysh.zerostate.my/install.sh | sh\n")
	}
}
