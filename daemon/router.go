package daemon

import (
	"crypto/tls"
	"fmt"
	"net/url"
	"sync"

	"github.com/PuvaanRaaj/devtun/config"
)

type route struct {
	target *url.URL
	cert   tls.Certificate
}

// Router holds domain → upstream route mappings and the TLS certificate map.
type Router struct {
	mu     sync.RWMutex
	routes map[string]*route // domain → route
	cfg    *config.Config
}

func NewRouter(cfg *config.Config) (*Router, error) {
	r := &Router{cfg: cfg}
	if err := r.build(cfg); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Router) build(cfg *config.Config) error {
	routes := make(map[string]*route, len(cfg.Domains))
	for _, d := range cfg.Domains {
		u, err := url.Parse(d.Target)
		if err != nil {
			return fmt.Errorf("invalid target %q for %s: %w", d.Target, d.Domain, err)
		}
		cert, err := tls.LoadX509KeyPair(
			certPath(cfg.Cert.CertDir, d.Domain),
			keyPath(cfg.Cert.CertDir, d.Domain),
		)
		if err != nil {
			return fmt.Errorf("load cert for %s: %w", d.Domain, err)
		}
		routes[d.Domain] = &route{target: u, cert: cert}
	}
	r.mu.Lock()
	r.routes = routes
	r.cfg = cfg
	r.mu.Unlock()
	return nil
}

// Reload atomically rebuilds routes from a new config.
func (r *Router) Reload(cfg *config.Config) error {
	return r.build(cfg)
}

// GetCertificate is used as tls.Config.GetCertificate for SNI-based routing.
func (r *Router) GetCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	route, ok := r.routes[hello.ServerName]
	if !ok {
		return nil, fmt.Errorf("no certificate for domain: %s", hello.ServerName)
	}
	c := route.cert
	return &c, nil
}

// Target returns the upstream URL for a given host, stripping any port suffix.
func (r *Router) Target(host string) (*url.URL, bool) {
	domain := stripPort(host)
	r.mu.RLock()
	defer r.mu.RUnlock()
	rt, ok := r.routes[domain]
	if !ok {
		return nil, false
	}
	return rt.target, true
}

// Domains returns all currently routed domains.
func (r *Router) Domains() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.routes))
	for d := range r.routes {
		out = append(out, d)
	}
	return out
}

func certPath(dir, domain string) string {
	return dir + "/" + domain + ".crt"
}

func keyPath(dir, domain string) string {
	return dir + "/" + domain + ".key"
}

func stripPort(host string) string {
	for i := len(host) - 1; i >= 0; i-- {
		if host[i] == ':' {
			return host[:i]
		}
		if host[i] == ']' {
			break
		}
	}
	return host
}
