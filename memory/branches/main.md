# Branch: main

## Session — 2026-03-09
Built proxysh from scratch and pushed to GitHub.

### What was built
- Full Go CLI (`cmd/`) with 8 commands: `start`, `stop`, `up`, `down`, `list`, `share`, `logs`, `doctor` + hidden `daemon`
- `cert/` — ECDSA P-256 CA generation, per-domain certs, macOS/Linux CA trust installation
- `daemon/` — TLS reverse proxy with SNI routing, WebSocket hijack, Unix socket IPC, hot-reload
- `hosts/` — `/etc/hosts` read/write via sudo
- `config/` — `.proxysh.yaml` loader with sane defaults
- `ipc/` — client/server JSON protocol over Unix socket
- `launchd/` — macOS LaunchAgent plist generation + `launchctl` load/unload
- `log/` — `log/slog` wrapper with file output support
- `docs/index.html` — single-file landing page, dark theme, install box with copy button
- `install.sh` — curl-pipe-sh installer targeting GitHub releases
- `Makefile` — build, install, release (cross-compile), test targets

### Status
- Compiles cleanly with `go build ./...`
- GitHub repo live: `PuvaanRaaj/proxysh`
- GitHub Pages live: `https://puvaanraaj.github.io/proxysh/`
- Relay server (`proxysh.show`) not yet implemented — `share` command has client-side code only
