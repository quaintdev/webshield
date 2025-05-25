package dot

import (
	"context"
	"crypto/tls"
	"github.com/miekg/dns"
	"github.com/quaintdev/webshield/src/internal/service"
	"log/slog"
	"sync"
	"time"
)

type Server struct {
	dnsService *service.DNSService
}

func NewDotServer(dnsService *service.DNSService) *Server {
	return &Server{
		dnsService: dnsService,
	}
}

func (s *Server) Start(ctx context.Context, wg *sync.WaitGroup, configService *service.ApplicationConfigService) {
	defer wg.Done()

	certConf := configService.GetCertConf()
	cert, err := tls.LoadX509KeyPair(certConf.CertPath, certConf.KeyPath)
	if err != nil {
		slog.Error("Failed to load TLS certificates", "error", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			slog.Info("Context cancelled, stopping DOT server")
			return
		default:
			// Continue with server setup
		}
		// Create a TLS configuration
		tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}
		baseListener, err := tls.Listen("tcp", ":853", tlsConfig)
		if err != nil {
			slog.Error("Failed to start TLS listener", "error", err)
			time.Sleep(5 * time.Second) // Wait before retry
			continue
		}

		// Use a function literal with defer for proper cleanup per iteration
		func() {
			defer baseListener.Close() // This will execute when this anonymous function returns

			// Wrap with tracking listener
			listener := &Listener{Listener: baseListener}
			handler := NewDNSHandler(s.dnsService, listener)

			slog.Info("DNS-over-TLS server started on port 853")
			// Create a DNS server
			dnsServer := &dns.Server{
				Listener:  listener,
				TLSConfig: tlsConfig,
				Handler:   dns.HandlerFunc(handler.HandleQuery(ctx)),
				Net:       "tcp-tls",
			}

			// Set up proper shutdown based on context
			go func() {
				<-ctx.Done()         // Wait for context cancellation
				dnsServer.Shutdown() // Graceful shutdown
			}()

			// Start the server
			if err := dnsServer.ActivateAndServe(); err != nil {
				slog.Error("DNS server stopped", "error", err)
			}
		}()

		// Only reach here after server has stopped
		slog.Info("Attempting to restart DNS-over-TLS server")
		time.Sleep(1 * time.Second) // Brief delay before restart
	}
}
