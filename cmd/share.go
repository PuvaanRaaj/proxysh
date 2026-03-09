package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"math/rand/v2"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var (
	sharePort     int
	shareRelay    string
	shareTTL      int
	sharePassword string
)

var shareCmd = &cobra.Command{
	Use:   "share",
	Short: "Share a local port via a public URL",
	Long: `Share creates a public HTTPS URL that tunnels traffic to your local port.

Examples:
  proxysh share --port 3000
  proxysh share --port 8080 --ttl 60`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if sharePort == 0 {
			return fmt.Errorf("--port is required")
		}

		subdomain := randomSubdomain()
		fmt.Printf("\nStarting tunnel for localhost:%d\n\n", sharePort)
		fmt.Printf("  Public URL: https://%s.%s\n\n", subdomain, shareDomain(shareRelay))
		fmt.Println("Forwarding requests... (press Ctrl+C to stop)")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		if shareTTL > 0 {
			go func() {
				time.Sleep(time.Duration(shareTTL) * time.Minute)
				fmt.Printf("\nTTL of %d minutes reached. Shutting down.\n", shareTTL)
				cancel()
			}()
		}

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigCh
			fmt.Println("\nTunnel closed.")
			cancel()
		}()

		return runTunnel(ctx, shareRelay, subdomain, sharePort, sharePassword)
	},
}

// runTunnel establishes a reverse tunnel to the relay server.
// The relay server accepts connections on subdomain.relay and forwards
// them through this tunnel to localhost:localPort.
func runTunnel(ctx context.Context, relay, subdomain string, localPort int, password string) error {
	relayAddr := relay + ":7000"
	tlsConf := &tls.Config{InsecureSkipVerify: true} //nolint:gosec

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 10 * time.Second}, "tcp", relayAddr, tlsConf)
		if err != nil {
			fmt.Printf("  Could not connect to relay (%s): %v\nRetrying in 5s...\n", relayAddr, err)
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(5 * time.Second):
				continue
			}
		}

		// Send registration
		req, _ := http.NewRequest("CONNECT", "/register", nil)
		req.Header.Set("X-Subdomain", subdomain)
		if password != "" {
			req.Header.Set("X-Password", password)
		}
		req.Write(conn) //nolint:errcheck

		fmt.Printf("  Tunnel active: https://%s.%s\n", subdomain, shareDomain(relay))

		// Forward connections
		tunnelForward(ctx, conn, localPort)
		conn.Close()

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(2 * time.Second):
		}
	}
}

func tunnelForward(ctx context.Context, tunnelConn net.Conn, localPort int) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		localConn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", localPort), 5*time.Second)
		if err != nil {
			return
		}

		go func() {
			defer localConn.Close()
			defer tunnelConn.Close()
			done := make(chan struct{}, 2)
			go func() { io.Copy(localConn, tunnelConn); done <- struct{}{} }() //nolint:errcheck
			go func() { io.Copy(tunnelConn, localConn); done <- struct{}{} }() //nolint:errcheck
			<-done
		}()
		return
	}
}

func randomSubdomain() string {
	adjectives := []string{"swift", "bright", "calm", "keen", "bold", "pure", "warm", "cool"}
	nouns := []string{"river", "cloud", "spark", "stone", "frost", "wind", "flame", "wave"}
	return adjectives[rand.IntN(len(adjectives))] + "-" + nouns[rand.IntN(len(nouns))] +
		fmt.Sprintf("-%d", rand.IntN(9000)+1000)
}

func shareDomain(relay string) string {
	if relay == "" || relay == "proxysh.show" {
		return "proxysh.show"
	}
	return relay
}

func init() {
	shareCmd.Flags().IntVarP(&sharePort, "port", "p", 0, "Local port to share (required)")
	shareCmd.Flags().StringVar(&shareRelay, "relay", "proxysh.show", "Relay server hostname")
	shareCmd.Flags().IntVar(&shareTTL, "ttl", 0, "Auto-expire tunnel after N minutes (0 = no limit)")
	shareCmd.Flags().StringVar(&sharePassword, "password", "", "Password protect the public URL")
	rootCmd.AddCommand(shareCmd)
}
