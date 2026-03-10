# devtun

Local HTTPS development domains & tunneling toolkit for macOS — written in Go.

Stop copy-pasting `localhost:3000`. Map real `.test` domains with valid HTTPS certs to your local services in seconds, and optionally share them publicly via a tunnel.

## Features

- **Real HTTPS locally** — generates a local CA, installs it into your system trust store, and issues per-domain certificates automatically
- **`.test` domains** — `https://example.test` instead of `http://localhost:3000`
- **Auto-start** — installs a LaunchAgent so the proxy daemon starts on every login
- **WebSocket support** — raw TCP tunneling for WS connections
- **Public sharing** — expose any local port via a temporary public URL through a relay server
- **Hot reload** — add/remove domains without restarting the daemon

## Requirements

- macOS (uses `pf` for port forwarding and `launchd` for auto-start)
- Go 1.25+

## Installation

```sh
curl -sL https://devtun.zerostate.my/install.sh | sh
```

Or with Homebrew:

```sh
brew install PuvaanRaaj/devtun/devtun
```

Or with Go:

```sh
go install github.com/PuvaanRaaj/devtun@latest
```

Or build from source:

```sh
git clone https://github.com/PuvaanRaaj/devtun
cd devtun
go build -o devtun .
```

## Quick Start

```sh
# First-time setup: generate CA, install trust, start daemon
devtun start

# Map a .test domain to a local port
devtun up example 3000
# → https://example.test now forwards to localhost:3000

# Add more domains
devtun up api 8080
devtun up admin 4000

# List active domains
devtun list

# Share a local port publicly (temporary URL)
devtun share --port 3000

# Health check
devtun doctor
```

## Commands

| Command | Description |
|---|---|
| `devtun start` | Generate CA, install trust store, start daemon |
| `devtun stop` | Stop the daemon |
| `devtun up <name> <port>` | Add `https://<name>.test → localhost:<port>` |
| `devtun down <name>` | Remove a domain mapping |
| `devtun list` | List all active domain mappings |
| `devtun share --port <n>` | Create a public URL tunneling to a local port |
| `devtun logs` | Tail the daemon log |
| `devtun doctor` | Run health checks |

### `devtun share` flags

| Flag | Default | Description |
|---|---|---|
| `--port`, `-p` | (required) | Local port to expose |
| `--relay` | `devtun.show` | Relay server hostname |
| `--ttl` | `0` (no limit) | Auto-close tunnel after N minutes |
| `--password` | — | Password-protect the public URL |

## How It Works

```
Browser → :443 → pf rdr → :8443 (daemon) → localhost:<port>
```

- The daemon listens on port **8443** (no root required at runtime)
- A `pf` redirect rule routes port **443 → 8443** (set up during `devtun start`)
- `/etc/hosts` entries point `.test` domains to `127.0.0.1`
- Each domain gets a TLS cert signed by the local CA
- The daemon hot-reloads its route table via a Unix socket at `/tmp/devtun.sock`

For public sharing, `devtun share` connects to a relay server over TLS and registers a random subdomain (e.g. `swift-river-4271.devtun.show`). Incoming connections are forwarded through the tunnel to your local port.

## File Locations

| Resource | Path |
|---|---|
| Config | `~/.devtun.yaml` |
| CA certificate | `~/.config/devtun/ca/ca.crt` |
| Domain certs | `~/.config/devtun/certs/<domain>.{crt,key}` |
| Daemon log | `~/.config/devtun/devtun.log` |
| LaunchAgent | `~/Library/LaunchAgents/com.PuvaanRaaj.devtun.plist` |
| IPC socket | `/tmp/devtun.sock` |

## Troubleshooting

Run `devtun doctor` first — it checks:

- CA certificate exists and is trusted
- Daemon is running
- LaunchAgent is installed
- Per-domain certificates and `/etc/hosts` entries
- `pf` port redirect is active

Most issues are fixed by running `devtun start` again.

**Port 443 redirect not working?** Run manually:
```sh
echo 'rdr pass on lo0 proto tcp from any to any port 443 -> 127.0.0.1 port 8443' | sudo pfctl -ef -
```

**Browser shows certificate warning?** The CA wasn't trusted. Re-run `devtun start` and approve the sudo prompt when installing the CA.

## License

MIT
