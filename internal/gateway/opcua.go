package gateway

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"strings"
	"time"

	v1 "github.com/simulator-bot/industrial-automation-connectors/internal/contracts/v1"
	"github.com/simulator-bot/industrial-automation-connectors/internal/equipment/model"
	"github.com/simulator-bot/industrial-automation-connectors/internal/platform/config"
	"github.com/simulator-bot/industrial-automation-connectors/internal/platform/httpserver"
	"github.com/simulator-bot/industrial-automation-connectors/internal/protocol"
	"github.com/simulator-bot/industrial-automation-connectors/internal/protocol/adapters/opcua"
)

type opcuaConfig struct {
	Endpoint       string                    `json:"endpoint"`
	ConnectionID   string                    `json:"connectionId"`
	DeviceID       string                    `json:"deviceId"`
	IntervalMillis int                       `json:"intervalMillis"`
	Deadband       float64                   `json:"deadband"`
	Points         []v1.OPCUAPoint           `json:"points"`
	Methods        map[string]v1.OPCUAMethod `json:"methods"`
}

type opcuaPilot struct {
	config    opcuaConfig
	gatewayID string
	session   protocol.Session
	targets   []string
	points    map[string]v1.OPCUAPoint
}

func openOPCUAPilot(gatewayConfig config.Config, path string, adapter protocol.Adapter) (*opcuaPilot, error) {
	if !httpserver.IsLoopbackAddress(gatewayConfig.HTTPAddress) {
		return nil, fmt.Errorf("GATEWAY_HTTP_ADDR must bind to a loopback IP when GATEWAY_OPCUA_CONFIG is set")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg opcuaConfig
	if err := decodeJSON(file, &cfg); err != nil {
		return nil, err
	}
	pilot, err := newOPCUAPilot(gatewayConfig.ServiceName, cfg)
	if err != nil {
		return nil, err
	}
	if err := pilot.connect(context.Background(), adapter); err != nil {
		return nil, err
	}
	return pilot, nil
}

func newOPCUAPilot(gatewayID string, cfg opcuaConfig) (*opcuaPilot, error) {
	if cfg.DeviceID == "" || cfg.ConnectionID == "" {
		return nil, fmt.Errorf("deviceId and connectionId are required")
	}
	if cfg.IntervalMillis < 0 || cfg.IntervalMillis > int(time.Hour/time.Millisecond) || cfg.Deadband < 0 {
		return nil, fmt.Errorf("intervalMillis must be between 0 and 3600000 and deadband must be nonnegative")
	}
	if len(cfg.Points) == 0 && len(cfg.Methods) == 0 {
		return nil, fmt.Errorf("at least one point or method is required")
	}
	pilot := &opcuaPilot{config: cfg, gatewayID: gatewayID, points: make(map[string]v1.OPCUAPoint)}
	pointIDs := make(map[string]bool)
	for _, point := range cfg.Points {
		if point.ID == "" || point.DataType == "" || pointIDs[point.ID] {
			return nil, fmt.Errorf("point IDs must be unique and each point must have a dataType")
		}
		node, err := opcua.ParseNodeID(point.NodeID)
		if err != nil {
			return nil, fmt.Errorf("invalid nodeId for point %q: %w", point.ID, err)
		}
		if _, exists := pilot.points[node.String()]; exists {
			return nil, fmt.Errorf("duplicate OPC UA NodeID %s", node)
		}
		pointIDs[point.ID] = true
		pilot.targets = append(pilot.targets, node.String())
		pilot.points[node.String()] = point
	}
	for name, method := range cfg.Methods {
		if name == "" || strings.Contains(name, "/") {
			return nil, fmt.Errorf("invalid method alias %q", name)
		}
		for _, target := range []string{method.ObjectNodeID, method.MethodNodeID} {
			if _, err := opcua.ParseNodeID(target); err != nil {
				return nil, fmt.Errorf("invalid NodeID for method %q: %w", name, err)
			}
		}
	}
	return pilot, nil
}

func (p *opcuaPilot) connect(ctx context.Context, adapter protocol.Adapter) error {
	var err error
	p.session, err = adapter.Connect(ctx, protocol.ConnectionConfig{
		Endpoint: p.config.Endpoint, ConnectionID: p.config.ConnectionID,
	})
	return err
}

func (p *opcuaPilot) subscribe(ctx context.Context) (<-chan protocol.Update, error) {
	return p.session.Subscribe(ctx, protocol.Subscription{
		Targets: p.targets, Mode: string(model.SubscriptionOnChange),
		Interval: time.Duration(p.config.IntervalMillis) * time.Millisecond, Deadband: p.config.Deadband,
	})
}

func (p *opcuaPilot) pointEvent(value protocol.Value) v1.PointEvent {
	point := p.points[value.TargetID]
	return v1.PointEvent{
		SchemaVersion: "1", EventID: rand.Text(), GatewayID: p.gatewayID,
		ConnectionID: p.config.ConnectionID, DeviceID: p.config.DeviceID,
		PointID: point.ID, Value: value.Value, DataType: point.DataType,
		Quality: value.Quality, ProtocolStatus: value.ProtocolStatus, Sequence: value.Sequence,
		SourceTimestamp: value.SourceTimestamp, IngestTimestamp: value.ReceivedAt,
	}
}

func (p *opcuaPilot) events(w http.ResponseWriter, r *http.Request) {
	if len(p.targets) == 0 {
		http.Error(w, "no OPC UA points configured", http.StatusNotFound)
		return
	}
	updates, err := p.subscribe(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-store")
	if err := writeSSE(w, "", nil); err != nil {
		return
	}

	keepalive := time.NewTicker(15 * time.Second)
	defer keepalive.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-keepalive.C:
			if err := writeSSE(w, "", nil); err != nil {
				return
			}
		case update, ok := <-updates:
			if !ok {
				return
			}
			if update.Err != nil {
				_ = writeSSE(w, "subscription-error", map[string]string{"error": update.Err.Error()})
				return
			}
			if err := writeSSE(w, "point", p.pointEvent(update.Value)); err != nil {
				return
			}
		}
	}
}

func writeSSE(w http.ResponseWriter, event string, value any) error {
	controller := http.NewResponseController(w)
	if err := controller.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return err
	}
	defer controller.SetWriteDeadline(time.Time{})
	if event == "" {
		if _, err := io.WriteString(w, ": keepalive\n\n"); err != nil {
			return err
		}
	} else {
		payload, err := json.Marshal(value)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, payload); err != nil {
			return err
		}
	}
	return controller.Flush()
}

func (p *opcuaPilot) invoke(w http.ResponseWriter, r *http.Request) {
	method, ok := p.config.Methods[r.PathValue("method")]
	if !ok {
		http.Error(w, "unknown OPC UA method alias", http.StatusNotFound)
		return
	}
	contentType, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if contentType != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}
	var input struct {
		Arguments []opcua.Argument `json:"arguments"`
	}
	if err := decodeJSON(http.MaxBytesReader(w, r.Body, 64<<10), &input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if input.Arguments == nil {
		http.Error(w, "arguments must be an array", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	result, err := p.callMethod(ctx, method, input.Arguments)
	if err != nil {
		status := http.StatusBadGateway
		if errors.Is(err, opcua.ErrInvalidArguments) {
			status = http.StatusBadRequest
		}
		http.Error(w, err.Error(), status)
		return
	}
	writeJSON(w, result)
}

func (p *opcuaPilot) callMethod(ctx context.Context, method v1.OPCUAMethod, args []v1.OPCUAArgument) (json.RawMessage, error) {
	payload, err := json.Marshal(opcua.MethodCall{ObjectNodeID: method.ObjectNodeID, Arguments: args})
	if err != nil {
		return nil, err
	}
	return p.session.Invoke(ctx, method.MethodNodeID, payload)
}

func decodeJSON(r io.Reader, dest any) error {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dest); err != nil {
		return err
	}
	if decoder.Decode(new(json.RawMessage)) != io.EOF {
		return fmt.Errorf("expected one JSON document")
	}
	return nil
}
