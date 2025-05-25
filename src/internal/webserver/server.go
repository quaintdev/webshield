package webserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"

	"github.com/quaintdev/webshield/src/internal/service"
)

type WebServer struct {
	dtMgmtService *service.DataMgmtService
	server        *http.Server
	dnsService    *service.DNSService
}

func NewWebServer(dtMgmtService *service.DataMgmtService, dnsService *service.DNSService) *WebServer {
	return &WebServer{
		dtMgmtService: dtMgmtService,
		dnsService:    dnsService,
	}
}

func (s *WebServer) Shutdown(ctx context.Context) {
	err := s.server.Shutdown(ctx)
	if err != nil {
		fmt.Println("Error shutting down server:", err)
		return
	}
}

func (s *WebServer) Start(wg *sync.WaitGroup) {
	defer wg.Done()

	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs)) // Serve static file+
	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/favicon.ico")
	})

	mux.HandleFunc("/guide", handleGuide)
	mux.HandleFunc("/", handleHome())

	mux.HandleFunc("POST /api/configurations", handleAddConfiguration(s.dtMgmtService))
	mux.HandleFunc("GET /api/configurations/{configId}", handleGetConfiguration(s.dtMgmtService))
	mux.HandleFunc("PUT /api/configurations/{configId}", handleUpdateConfiguration(s.dtMgmtService))
	mux.HandleFunc("DELETE /api/configurations/{configId}", handleDeleteConfiguration(s.dtMgmtService))
	mux.HandleFunc("POST /api/configurations/{configId}/state", handleConfigurationState(s.dtMgmtService))
	mux.HandleFunc("GET /api/configurations", handleGetConfigurations(s.dtMgmtService))

	//DoH Server
	mux.HandleFunc("/doh/{configId}", handleDoHQuery(s.dnsService))
	port := os.Getenv("PORT")
	s.server = &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	slog.Info("WebServer started", "port", port)
	err := s.server.ListenAndServe()
	if err != nil {
		fmt.Println("error starting server", err)
		return
	}
}
