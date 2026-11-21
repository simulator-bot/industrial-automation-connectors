package process

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/simulator-bot/industrial-automation-connectors/internal/equipment/repository"
	"github.com/simulator-bot/industrial-automation-connectors/internal/platform/config"
	"github.com/simulator-bot/industrial-automation-connectors/internal/platform/httpserver"
)

type Server struct {
	http *httpserver.Server
}

type Options struct {
	CatalogPath string
	GatewayURL  string
}

func NewServer(cfg config.Config, logger *slog.Logger, opts Options) (*Server, error) {
	mux := http.NewServeMux()
	local := httpserver.IsLoopbackAddress(cfg.HTTPAddress)
	if !local && (opts.CatalogPath != "" || opts.GatewayURL != "") {
		return nil, fmt.Errorf("the equipment workbench requires PROCESS_HTTP_ADDR to bind to a loopback IP")
	}
	catalog, err := repository.OpenCatalog(opts.CatalogPath)
	if err != nil {
		return nil, err
	}
	if err := validateCatalog(catalog.Snapshot()); err != nil {
		return nil, fmt.Errorf("invalid equipment catalog: %w", err)
	}
	if local {
		api, err := newCatalogAPI(catalog, opts.GatewayURL)
		if err != nil {
			return nil, err
		}
		mux.HandleFunc("GET /api/v1/equipment/configuration", httpserver.LocalOnly(api.get))
		mux.HandleFunc("PUT /api/v1/equipment/configuration", httpserver.LocalOnly(api.put))
		mux.HandleFunc("POST /api/v1/equipment/instances/{id}/test", httpserver.LocalOnly(api.test))
	}
	mux.HandleFunc("GET /healthz", health(cfg.ServiceName))
	mux.HandleFunc("GET /readyz", health(cfg.ServiceName))
	mux.HandleFunc("GET /api/v1/equipment", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"items": catalog.Snapshot().Instances})
	})
	return &Server{http: httpserver.New(cfg.ServiceName, cfg.HTTPAddress, mux, logger)}, nil
}

func (s *Server) Run() error { return s.http.Run() }

func health(service string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": service})
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
