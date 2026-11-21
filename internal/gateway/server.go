package gateway

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/simulator-bot/industrial-automation-connectors/internal/platform/config"
	"github.com/simulator-bot/industrial-automation-connectors/internal/platform/httpserver"
	"github.com/simulator-bot/industrial-automation-connectors/internal/protocol"
	"github.com/simulator-bot/industrial-automation-connectors/internal/protocol/adapters/opcua"
)

type Server struct {
	http  *httpserver.Server
	pilot *opcuaPilot
}

func NewServer(cfg config.Config, logger *slog.Logger, opcuaConfigPath string) (*Server, error) {
	registry := protocol.NewRegistry()
	if err := registry.Register(opcua.Adapter{}); err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", health(cfg.ServiceName))
	adapter, _ := registry.Get(opcua.Name)
	if httpserver.IsLoopbackAddress(cfg.HTTPAddress) {
		tests := opcuaTests{gatewayID: cfg.ServiceName, adapter: adapter, slots: make(chan struct{}, 4)}
		mux.HandleFunc("POST /api/v1/gateway/opcua/test", httpserver.LocalOnly(tests.handle))
	}
	var pilot *opcuaPilot
	if opcuaConfigPath != "" {
		var err error
		pilot, err = openOPCUAPilot(cfg, opcuaConfigPath, adapter)
		if err != nil {
			return nil, fmt.Errorf("configure OPC UA pilot: %w", err)
		}
		mux.HandleFunc("GET /api/v1/gateway/opcua/events", httpserver.LocalOnly(pilot.events))
		mux.HandleFunc("POST /api/v1/gateway/opcua/methods/{method}", httpserver.LocalOnly(pilot.invoke))
	}
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if pilot != nil && pilot.session.Health(r.Context()) != nil {
			http.Error(w, "OPC UA session unavailable", http.StatusServiceUnavailable)
			return
		}
		health(cfg.ServiceName)(w, r)
	})
	mux.HandleFunc("GET /api/v1/gateway", func(w http.ResponseWriter, r *http.Request) {
		status := map[string]any{
			"service":          cfg.ServiceName,
			"state":            "idle",
			"protocolAdapters": registry.Descriptors(),
		}
		if pilot != nil {
			status["state"] = "connected"
			if pilot.session.Health(r.Context()) != nil {
				status["state"] = "disconnected"
			}
		}
		writeJSON(w, status)
	})
	return &Server{http: httpserver.New(cfg.ServiceName, cfg.HTTPAddress, mux, logger), pilot: pilot}, nil
}

func (s *Server) Run() error {
	if s.pilot != nil {
		defer s.pilot.session.Close()
	}
	return s.http.Run()
}

func health(service string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]string{"status": "ok", "service": service})
	}
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}
