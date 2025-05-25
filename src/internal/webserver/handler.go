package webserver

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/miekg/dns"
	"github.com/quaintdev/webshield/src/internal/apperrors"
	"github.com/quaintdev/webshield/src/internal/dto"
	"github.com/quaintdev/webshield/src/internal/service"
)

func handleHome() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/home.html")
	}
}

func handleAddConfiguration(service *service.DataMgmtService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var req *dto.AddPresetRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			slog.Error("Failed to decode configuration: ", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		response, err := service.AddConfig(r.Context(), req)
		if err != nil {
			m := struct {
				Message string `json:"message"`
			}{
				Message: err.Error(),
			}
			if errors.Is(err, apperrors.ErrNoSubscription) ||
				errors.Is(err, apperrors.ErrMaxConfigsReached) {
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(m)
				return
			} else {
				slog.Error("Failed to add config: ", "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		json.NewEncoder(w).Encode(response)
	}
}

func handleGetConfiguration(service *service.DataMgmtService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		configId := r.PathValue("configId")
		slog.Debug("GET configuration ", "configId", configId)
		presetResponse, err := service.GetConfig(r.Context(), configId)
		if err != nil {
			slog.Error("Failed to get config: ", "error", err)
			return
		}
		json.NewEncoder(w).Encode(presetResponse)
	}
}

func handleUpdateConfiguration(service *service.DataMgmtService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		configId := r.PathValue("configId")
		slog.Debug("PUT config request received", "configId", configId)
		req := dto.UpdatePresetRequest{}
		json.NewDecoder(r.Body).Decode(&req)
		req.PresetID = configId
		presetResponse, err := service.UpdateConfig(r.Context(), &req)
		if err != nil {
			return
		}
		json.NewEncoder(w).Encode(presetResponse)
	}
}

func handleDeleteConfiguration(service *service.DataMgmtService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		configId := r.PathValue("configId")
		slog.Debug("DELETE config request received", "configId", configId)
		err := service.DeleteConfig(r.Context(), configId)
		if err != nil {
			slog.Error("Failed to delete config: ", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleGetConfigurations(us *service.DataMgmtService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		configsFromDB, err := us.GetAllConfigs(r.Context())
		if err != nil {
			slog.Error("error fetching configurations: ", "us.GetAllConfigs", err)
			return
		}
		configs := make([]*dto.PresetResponse, len(configsFromDB))
		for i, dbConfig := range configsFromDB {
			configs[i] = dto.MakePresetResponse(dbConfig) // Assign directly to index
		}
		json.NewEncoder(w).Encode(configs)
	}
}

func handleConfigurationState(service *service.DataMgmtService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		configId := r.PathValue("configId")
		req := struct {
			Enabled bool `json:"enabled"`
		}{}
		json.NewDecoder(r.Body).Decode(&req)
		err := service.SetConfigState(r.Context(), configId, req.Enabled)
		if err != nil {
			slog.Error("Failed to get config state: ", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func handleGuide(w http.ResponseWriter, r *http.Request) {

	configId := r.URL.Query().Get("configId")
	name := r.URL.Query().Get("name")

	tmplPath := "static/guide.html"

	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		slog.Error("failed to parse template: ", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = tmpl.Execute(w, struct {
		ConfigId string
		Name     string
		Host     string
	}{
		ConfigId: configId,
		Name:     name,
		Host:     os.Getenv("hostname"),
	})
	if err != nil {
		slog.Error("failed to execute template: ", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func handleDoHQuery(service *service.DNSService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only accept GET and POST methods
		if r.Method != http.MethodGet && r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var (
			buf []byte
			err error
		)
		configId := r.PathValue("configId")
		switch r.Method {
		case http.MethodGet:
			dnsParam := r.URL.Query().Get("dns")
			if dnsParam == "" {
				http.Error(w, "Missing 'dns' parameter", http.StatusBadRequest)
				return
			}

			// Decode the base64url encoded DNS message
			buf, err = base64.RawURLEncoding.DecodeString(dnsParam)
			if err != nil {
				http.Error(w, "Invalid DNS parameter encoding", http.StatusBadRequest)
				return
			}
		case http.MethodPost:
			contentType := r.Header.Get("Content-Type")
			if contentType != "application/dns-message" {
				http.Error(w, "Unsupported Content-Type", http.StatusUnsupportedMediaType)
				return
			}

			// Read the DNS message from request body
			buf, err = io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Failed to read request body", http.StatusBadRequest)
				return
			}
		}

		// Parse the DNS message
		msg := new(dns.Msg)
		if err := msg.Unpack(buf); err != nil {
			http.Error(w, "Failed to parse DNS message", http.StatusBadRequest)
			return
		}

		response, err := service.ProcessQuery(r.Context(), msg, configId)
		if err != nil {
			return
		}
		// Pack the response into wire format
		respBytes, err := response.Pack()
		if err != nil {
			http.Error(w, "Failed to encode DNS response", http.StatusInternalServerError)
			return
		}
		// Set appropriate headers and write response
		w.Header().Set("Content-Type", "application/dns-message")
		w.Header().Set("Content-Length", strconv.Itoa(len(respBytes)))
		w.Header().Set("Cache-Control", getCacheHeader(response))
		w.WriteHeader(http.StatusOK)
		w.Write(respBytes)
	}
}

// getCacheHeader determines the Cache-Control header based on DNS response
func getCacheHeader(msg *dns.Msg) string {
	// Find the lowest TTL in the response
	minTTL := uint32(3600) // Default to 1 hour

	// Check all record types that have TTL
	for _, section := range [][]dns.RR{msg.Answer, msg.Ns, msg.Extra} {
		for _, rr := range section {
			if rr.Header().Ttl < minTTL {
				minTTL = rr.Header().Ttl
			}
		}
	}

	// Safeguard against very short TTLs
	if minTTL < 10 {
		minTTL = 10
	}

	//return fmt.Sprintf("max-age=%d, must-revalidated", minTTL)
	return fmt.Sprintf("max-age=%d", minTTL)
}
