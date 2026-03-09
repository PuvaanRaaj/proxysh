package daemon

import (
	"encoding/json"
	"net"
	"os"

	"github.com/PuvaanRaaj/proxysh/config"
	"github.com/PuvaanRaaj/proxysh/ipc"
	proxylog "github.com/PuvaanRaaj/proxysh/log"
)

// ServeIPC listens on a Unix socket and handles CLI → daemon commands.
func ServeIPC(socketPath string, router *Router, cfgPath string, shutdown chan<- struct{}) {
	os.Remove(socketPath)
	if err := os.MkdirAll(parentDir(socketPath), 0755); err != nil {
		proxylog.Error("ipc mkdir", "err", err)
		return
	}

	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		proxylog.Error("ipc listen", "err", err)
		return
	}
	defer ln.Close()
	proxylog.Info("ipc listening", "socket", socketPath)

	for {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		go handleIPCConn(conn, router, cfgPath, shutdown)
	}
}

func handleIPCConn(conn net.Conn, router *Router, cfgPath string, shutdown chan<- struct{}) {
	defer conn.Close()
	enc := json.NewEncoder(conn)

	var req ipc.Request
	if err := json.NewDecoder(conn).Decode(&req); err != nil {
		enc.Encode(&ipc.Response{OK: false, Error: err.Error()}) //nolint:errcheck
		return
	}

	switch req.Cmd {
	case ipc.CmdReload:
		cfg, err := config.Load(cfgPath)
		if err != nil {
			enc.Encode(&ipc.Response{OK: false, Error: err.Error()}) //nolint:errcheck
			return
		}
		if err := router.Reload(cfg); err != nil {
			enc.Encode(&ipc.Response{OK: false, Error: err.Error()}) //nolint:errcheck
			return
		}
		proxylog.Info("config reloaded")
		enc.Encode(&ipc.Response{OK: true}) //nolint:errcheck

	case ipc.CmdStatus:
		domains := router.Domains()
		statuses := make([]ipc.DomainStatus, 0, len(domains))
		for _, d := range domains {
			target, _ := router.Target(d)
			statuses = append(statuses, ipc.DomainStatus{
				Domain: d,
				Target: target.String(),
				Active: true,
			})
		}
		enc.Encode(&ipc.Response{OK: true, Domains: statuses}) //nolint:errcheck

	case ipc.CmdShutdown:
		enc.Encode(&ipc.Response{OK: true}) //nolint:errcheck
		shutdown <- struct{}{}

	default:
		enc.Encode(&ipc.Response{OK: false, Error: "unknown command: " + req.Cmd}) //nolint:errcheck
	}
}

func parentDir(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[:i]
		}
	}
	return "."
}
