package daemon

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/PuvaanRaaj/proxysh/config"
	proxylog "github.com/PuvaanRaaj/proxysh/log"
)

// Run starts the daemon: TLS proxy server + IPC socket server.
// cfgPath is the path to .proxysh.yaml, used for hot-reloads.
func Run(cfg *config.Config, cfgPath string) error {
	if cfg.Daemon.LogFile != "" {
		if err := proxylog.SetFile(cfg.Daemon.LogFile); err != nil {
			proxylog.Warn("could not open log file, using stderr", "err", err)
		}
	}

	router, err := NewRouter(cfg)
	if err != nil {
		return fmt.Errorf("build router: %w", err)
	}

	tlsCfg := &tls.Config{
		GetCertificate: router.GetCertificate,
		MinVersion:     tls.VersionTLS12,
	}

	addr := fmt.Sprintf("127.0.0.1:%d", cfg.Daemon.ListenPort)
	ln, err := tls.Listen("tcp", addr, tlsCfg)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}

	srv := &http.Server{
		Handler: Handler(router),
	}

	// Also serve plain HTTP on port 80-equivalent to redirect to HTTPS
	httpPort := 8080
	httpLn, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", httpPort))
	if err != nil {
		proxylog.Warn("could not bind HTTP redirect port", "port", httpPort, "err", err)
	}

	shutdown := make(chan struct{}, 1)

	// IPC server
	go ServeIPC(config.IPCSocketPath, router, cfgPath, shutdown)

	// HTTP → HTTPS redirect
	if httpLn != nil {
		httpSrv := &http.Server{
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				target := "https://" + r.Host + r.URL.RequestURI()
				http.Redirect(w, r, target, http.StatusMovedPermanently)
			}),
		}
		go httpSrv.Serve(httpLn) //nolint:errcheck
	}

	// Signal handling
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
		for sig := range sigCh {
			switch sig {
			case syscall.SIGHUP:
				newCfg, err := config.Load(cfgPath)
				if err != nil {
					proxylog.Error("reload config", "err", err)
					continue
				}
				if err := router.Reload(newCfg); err != nil {
					proxylog.Error("reload router", "err", err)
					continue
				}
				proxylog.Info("reloaded config via SIGHUP")
			case syscall.SIGTERM, syscall.SIGINT:
				shutdown <- struct{}{}
			}
		}
	}()

	proxylog.Info("proxysh daemon started", "addr", addr)

	// Write PID
	writePID(cfg.Daemon.PIDFile)

	go func() {
		<-shutdown
		proxylog.Info("shutting down")
		srv.Shutdown(context.Background()) //nolint:errcheck
		os.Remove(cfg.Daemon.PIDFile)
		os.Remove(config.IPCSocketPath)
	}()

	if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func writePID(path string) {
	if path == "" {
		return
	}
	os.MkdirAll(parentDir(path), 0755) //nolint:errcheck
	os.WriteFile(path, []byte(fmt.Sprintf("%d\n", os.Getpid())), 0644) //nolint:errcheck
}
