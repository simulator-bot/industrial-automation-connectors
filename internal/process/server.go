package process

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/simulator-bot/industrial-automation-connectors/internal/equipment/repository"
	"github.com/simulator-bot/industrial-automation-connectors/internal/platform/config"
	"github.com/simulator-bot/industrial-automation-connectors/internal/platform/httpserver"
)

type Server struct {
	http *httpserver.Server
}

func NewServer(cfg config.Config, logger *slog.Logger) *Server {
	mux := http.NewServeMux()
	devices := repository.NewMemoryDevices()
	mux.HandleFunc("GET /healthz", health(cfg.ServiceName))
	mux.HandleFunc("GET /readyz", health(cfg.ServiceName))
	mux.HandleFunc("GET /api/v1/equipment", func(w http.ResponseWriter, r *http.Request) {
		items, err := devices.List(r.Context())
		if err != nil {
			http.Error(w, "equipment query failed", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	})
	return &Server{http: httpserver.New(cfg.ServiceName, cfg.HTTPAddress, mux, logger)}
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
