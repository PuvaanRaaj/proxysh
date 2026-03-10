# Services & Paths

## GitHub
- Repo: `PuvaanRaaj/devtun` (public, `main` branch)
- GitHub Pages: `https://puvaanraaj.github.io/devtun/` (from `/docs` on `main`)

## Runtime Paths
| Resource | Path |
|---|---|
| CA certificate | `~/.config/devtun/ca/ca.crt` |
| CA private key | `~/.config/devtun/ca/ca.key` |
| Domain certs | `~/.config/devtun/certs/<domain>.{crt,key}` |
| Daemon log | `~/.config/devtun/devtun.log` |
| PID file | `~/.config/devtun/devtun.pid` |
| IPC socket | `/tmp/devtun.sock` |
| LaunchAgent plist | `~/Library/LaunchAgents/com.PuvaanRaaj.devtun.plist` |
| User config | `~/.devtun.yaml` (or nearest `.devtun.yaml` walking up from cwd) |

## External Services
- Relay server: `devtun.show:7000` (TLS) — used by `devtun share`; client-side only, relay server not yet implemented
