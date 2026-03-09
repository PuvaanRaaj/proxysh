package daemon

// Package daemon implements the background proxy server process.
// It is started by the CLI via `proxysh daemon --config <path>` and
// runs until stopped via IPC or signal.
