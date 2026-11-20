package gateway

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/simulator-bot/industrial-automation-connectors/internal/platform/config"
	"github.com/simulator-bot/industrial-automation-connectors/internal/platform/httpserver"
	"github.com/simulator-bot/industrial-automation-connectors/internal/protocol"
)

type Server struct {
	http *httpserver.Server
}

func NewServer(cfg config.Config, logger *slog.Logger) *Server {
	registry := protocol.NewRegistry()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", health(cfg.ServiceName))
	mux.HandleFunc("GET /readyz", health(cfg.ServiceName))
	mux.HandleFunc("GET /api/v1/gateway", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"service":          cfg.ServiceName,
			"state":            "idle",
			"protocolAdapters": registry.Descriptors(),
		})
	})
	return &Server{http: httpserver.New(cfg.ServiceName, cfg.HTTPAddress, mux, logger)}
}

func (s *Server) Run() error { return s.http.Run() }

func health(service string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "service": service})
	}
}
