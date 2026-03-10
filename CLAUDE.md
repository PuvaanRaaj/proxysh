# devtun

Local HTTPS development domains & tunneling toolkit — written in Go.

## Quick Reference
```
devtun start              # generate CA, install trust, start daemon
devtun up myapp 3000      # https://myapp.test → localhost:3000
devtun list               # list active domains
devtun share --port 3000  # public URL via tunnel
devtun doctor             # health checks
```

## Codebase Memory
> Auto-generated — edit /memory/*.md, not this file.
> Last synced: 2026-03-09

---

### Conventions

**Go**
- Module: `github.com/PuvaanRaaj/devtun`
- Go version: 1.25 (always use latest)
- CLI framework: `github.com/spf13/cobra`
- Config format: YAML via `gopkg.in/yaml.v3`
- Cert algorithm: ECDSA P-256 (not RSA — smaller, faster, modern)
- IPC transport: Unix socket at `/tmp/devtun.sock`, encoded as JSON

**Architecture**
- Daemon binds to port 8443; pf rdr rule redirects 443 → 8443 (no root needed at runtime)
- LaunchAgent (not LaunchDaemon) — runs as user, auto-starts on login
- Hot reload: CLI sends `{"cmd":"reload"}` over Unix socket; daemon swaps route table atomically under `sync.RWMutex`

**Static Site**
- Landing page: `docs/index.html` (single file, no build step)
- GitHub Pages from `/docs` on `main` — `https://puvaanraaj.github.io/devtun/`

---

### Services & Paths

| Resource | Path |
|---|---|
| CA certificate | `~/.config/devtun/ca/ca.crt` |
| Domain certs | `~/.config/devtun/certs/<domain>.{crt,key}` |
| Daemon log | `~/.config/devtun/devtun.log` |
| IPC socket | `/tmp/devtun.sock` |
| LaunchAgent plist | `~/Library/LaunchAgents/com.PuvaanRaaj.devtun.plist` |
| Relay server | `devtun.show:7000` (TLS, client-side only — not yet hosted) |

---

### Gotchas

- Port 443 requires root on macOS — daemon binds to 8443, pf rdr rule redirects 443 → 8443
- `httputil.ReverseProxy` does NOT handle WebSocket — must hijack conn and `io.Copy` raw TCP
- `gh repo create --source=.` requires `git init` first
- macOS cert max validity: 825 days
- ECDSA PEM: use `x509.MarshalECPrivateKey` / `x509.ParseECPrivateKey` (not PKCS8)
