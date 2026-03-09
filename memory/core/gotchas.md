# Gotchas

## macOS / System
- Port 443 requires root on macOS — daemon binds to 8443, pf rdr rule redirects 443 → 8443
- CA trust requires one sudo: `sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain ca.crt`
- `/etc/hosts` edits require sudo — use `sudo sh -c 'echo "..." >> /etc/hosts'`
- `gh repo create --source=.` requires `git init` first or it fails with "not a git repository"

## Proxy / TLS
- `httputil.ReverseProxy` does NOT handle WebSocket upgrades — must detect `Upgrade: websocket` header, hijack the conn via `http.Hijacker`, dial upstream directly, and `io.Copy` in both directions
- `tls.Listen` with `GetCertificate` callback enables SNI routing — one listener, multiple domains, certs loaded lazily per domain name

## Certificates
- macOS trusts certs for max 825 days — `days_valid: 825` in config
- ECDSA keys: use `x509.MarshalECPrivateKey` / `x509.ParseECPrivateKey` (not the generic PKCS8 variants) for PEM encoding
