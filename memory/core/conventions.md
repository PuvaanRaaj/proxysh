# Conventions

## Go
- Module: `github.com/PuvaanRaaj/devtun`
- Go version: 1.25 (always use latest)
- CLI framework: `github.com/spf13/cobra`
- Config format: YAML via `gopkg.in/yaml.v3`
- Cert algorithm: ECDSA P-256 (not RSA — smaller, faster, modern)
- IPC transport: Unix socket at `/tmp/devtun.sock`, encoded as JSON

## Architecture
- Daemon binds to port 8443; pf rdr rule redirects 443 → 8443 (no root needed at runtime)
- LaunchAgent (not LaunchDaemon) — runs as user, auto-starts on login
- LaunchAgent plist: `~/Library/LaunchAgents/com.PuvaanRaaj.devtun.plist`
- Hot reload: CLI sends `{"cmd":"reload"}` over Unix socket; daemon swaps route table atomically under `sync.RWMutex`

## Static Site
- Landing page: `docs/index.html` (single file, no build step)
- Served via GitHub Pages from `/docs` on `main` branch
- Install section uses a terminal-style box with JS clipboard copy button
