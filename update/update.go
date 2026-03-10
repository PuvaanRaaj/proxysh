package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const releaseAPI = "https://api.github.com/repos/PuvaanRaaj/devtun/releases/latest"

// CheckAsync starts a background update check and returns a function that
// waits for the result (up to the HTTP timeout) and prints a notice if a
// newer version is available. Call the returned function after the command runs.
func CheckAsync(current string) func() {
	if current == "dev" || current == "" {
		return func() {}
	}

	type result struct{ msg string }
	ch := make(chan result, 1)

	go func() {
		client := &http.Client{Timeout: 3 * time.Second}
		resp, err := client.Get(releaseAPI)
		if err != nil {
			ch <- result{}
			return
		}
		defer resp.Body.Close()

		var release struct {
			TagName string `json:"tag_name"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
			ch <- result{}
			return
		}

		latest := strings.TrimPrefix(release.TagName, "v")
		curr := strings.TrimPrefix(current, "v")

		if latest != "" && latest != curr {
			ch <- result{msg: fmt.Sprintf(
				"\n  A new version of devtun is available: v%s → v%s\n  Update: curl -sL https://devtun.zerostate.my/install.sh | sh\n",
				curr, latest,
			)}
		} else {
			ch <- result{}
		}
	}()

	return func() {
		select {
		case r := <-ch:
			if r.msg != "" {
				fmt.Print(r.msg)
			}
		case <-time.After(3 * time.Second):
		}
	}
}
