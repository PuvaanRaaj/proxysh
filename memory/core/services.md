# Services & Paths

## GitHub
- Repo: `PuvaanRaaj/proxysh` (public, `main` branch)
- GitHub Pages: `https://puvaanraaj.github.io/proxysh/` (from `/docs` on `main`)

## Runtime Paths
| Resource | Path |
|---|---|
| CA certificate | `~/.config/proxysh/ca/ca.crt` |
| CA private key | `~/.config/proxysh/ca/ca.key` |
| Domain certs | `~/.config/proxysh/certs/<domain>.{crt,key}` |
| Daemon log | `~/.config/proxysh/proxysh.log` |
| PID file | `~/.config/proxysh/proxysh.pid` |
| IPC socket | `/tmp/proxysh.sock` |
| LaunchAgent plist | `~/Library/LaunchAgents/com.PuvaanRaaj.proxysh.plist` |
| User config | `~/.proxysh.yaml` (or nearest `.proxysh.yaml` walking up from cwd) |

## External Services
- Relay server: `proxysh.show:7000` (TLS) — used by `proxysh share`; client-side only, relay server not yet implemented
