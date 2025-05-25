package dot

import (
	"context"
	"log/slog"

	"github.com/quaintdev/webshield/src/internal/service"

	"github.com/miekg/dns"
)

type DNSHandler struct {
	dnsService *service.DNSService
	listener   *Listener
}

func NewDNSHandler(dnsService *service.DNSService, listener *Listener) *DNSHandler {
	return &DNSHandler{
		dnsService: dnsService,
		listener:   listener,
	}
}

func (h *DNSHandler) HandleQuery(ctx context.Context) func(w dns.ResponseWriter, r *dns.Msg) {
	return func(w dns.ResponseWriter, r *dns.Msg) {
		var serverName string
		if sni, ok := h.listener.GetServerName(w.RemoteAddr().String()); ok {
			slog.Debug("fetched server name", "server", sni)
			serverName = sni
		}
		response, err := h.dnsService.ProcessQuery(ctx, r, serverName)
		if err != nil {
			slog.Error("error occurred while processing request")
			return
		}
		// Write the response back to the client
		if err := w.WriteMsg(response); err != nil {
			slog.Error("failed to write DNS response", "error", err)
		}
	}
}
