package daemon

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"time"

	tunlog "github.com/PuvaanRaaj/devtun/log"
)

// Handler returns an http.Handler that reverse-proxies to the correct upstream
// based on the Host header, using the Router for routing.
func Handler(router *Router) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		target, ok := router.Target(r.Host)
		if !ok {
			http.Error(w, fmt.Sprintf("no route for host: %s", r.Host), http.StatusBadGateway)
			return
		}

		tunlog.Info("proxy", "method", r.Method, "host", r.Host, "path", r.URL.Path)

		if isWebSocketUpgrade(r) {
			proxyWebSocket(w, r, target.Host)
			return
		}

		rp := &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.URL.Scheme = target.Scheme
				req.URL.Host = target.Host
				req.Host = target.Host
				if target.Path != "" {
					req.URL.Path = target.Path + req.URL.Path
				}
			},
			ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
				tunlog.Error("upstream error", "err", err, "host", r.Host)
				http.Error(w, "upstream unavailable: "+err.Error(), http.StatusBadGateway)
			},
			ModifyResponse: func(resp *http.Response) error {
				resp.Header.Set("X-Proxied-By", "devtun")
				return nil
			},
		}
		rp.ServeHTTP(w, r)
	})
}

func isWebSocketUpgrade(r *http.Request) bool {
	return r.Header.Get("Upgrade") == "websocket"
}

func proxyWebSocket(w http.ResponseWriter, r *http.Request, upstreamAddr string) {
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "websocket not supported", http.StatusInternalServerError)
		return
	}

	upstream, err := net.DialTimeout("tcp", upstreamAddr, 10*time.Second)
	if err != nil {
		http.Error(w, "upstream connection failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer upstream.Close()

	if err := r.Write(upstream); err != nil {
		http.Error(w, "failed to forward request", http.StatusBadGateway)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		return
	}
	defer clientConn.Close()

	done := make(chan struct{}, 2)
	go func() {
		io.Copy(upstream, clientConn) //nolint:errcheck
		done <- struct{}{}
	}()
	go func() {
		io.Copy(clientConn, upstream) //nolint:errcheck
		done <- struct{}{}
	}()
	<-done
}
