package opcua

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
	"github.com/simulator-bot/industrial-automation-connectors/internal/equipment/model"
	"github.com/simulator-bot/industrial-automation-connectors/internal/protocol"
)

const Name = "opcua"
const requestTimeout = 5 * time.Second

type Adapter struct{}

func (Adapter) Descriptor() protocol.Descriptor {
	return protocol.Descriptor{
		Protocol: Name,
		Version:  "1",
		Features: []string{"read", "subscribe", "invoke"},
	}
}

func (Adapter) Connect(ctx context.Context, cfg protocol.ConnectionConfig) (protocol.Session, error) {
	endpoint, err := SimulatorEndpoint(cfg.Endpoint)
	if err != nil {
		return nil, err
	}
	if cfg.Credentials != "" || len(cfg.Options) != 0 {
		return nil, fmt.Errorf("OPC UA simulator connections support anonymous authentication only")
	}

	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	endpoints, err := opcua.GetEndpoints(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("discover OPC UA endpoints: %w", err)
	}
	for _, ep := range endpoints {
		if ep.SecurityMode != ua.MessageSecurityModeNone || ep.SecurityPolicyURI != ua.SecurityPolicyURINone {
			continue
		}
		for _, token := range ep.UserIdentityTokens {
			if token.TokenType == ua.UserTokenTypeAnonymous {
				return connect(ctx, endpoint, ep)
			}
		}
	}
	return nil, fmt.Errorf("OPC UA simulator has no anonymous endpoint with SecurityPolicy None")
}

func SimulatorEndpoint(raw string) (string, error) {
	endpoint, err := url.Parse(raw)
	if err != nil || endpoint.Scheme != "opc.tcp" || endpoint.Hostname() == "" ||
		endpoint.User != nil || endpoint.ForceQuery || endpoint.RawQuery != "" || endpoint.Fragment != "" {
		return "", fmt.Errorf("expected an opc.tcp simulator endpoint")
	}
	host := endpoint.Hostname()
	if strings.EqualFold(host, "localhost") {
		// Local simulators commonly bind IPv4 only, while the UA dialer uses
		// just the first resolved address for a hostname.
		port := endpoint.Port()
		if port == "" {
			port = "4840"
		}
		endpoint.Host = net.JoinHostPort("127.0.0.1", port)
	} else if !net.ParseIP(host).IsLoopback() {
		return "", fmt.Errorf("OPC UA SecurityPolicy None is limited to loopback simulator endpoints")
	}
	return endpoint.String(), nil
}

func connect(ctx context.Context, endpoint string, ep *ua.EndpointDescription) (*session, error) {
	disconnected := make(chan struct{})
	client, err := opcua.NewClient(endpoint,
		opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeAnonymous),
		opcua.RequestTimeout(requestTimeout),
		// gopcua v0.9.1 can race with Close while replacing its connection.
		opcua.AutoReconnect(false),
		opcua.StateChangedFunc(func(state opcua.ConnState) {
			if state == opcua.Disconnected {
				close(disconnected)
			}
		}),
	)
	if err != nil {
		return nil, err
	}
	if err := client.Connect(ctx); err != nil {
		_ = client.Close(ctx)
		return nil, fmt.Errorf("connect to OPC UA simulator: %w", err)
	}
	return &session{client: client, done: make(chan struct{}), disconnected: disconnected}, nil
}

type session struct {
	client       *opcua.Client
	done         chan struct{}
	disconnected <-chan struct{}
	once         sync.Once
}

func (s *session) Health(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if state := s.client.State(); state != opcua.Connected {
		return fmt.Errorf("OPC UA connection is %v", state)
	}
	return nil
}

func (s *session) Close() error {
	s.once.Do(func() {
		close(s.done)
		ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
		defer cancel()
		_ = s.client.Close(ctx)
	})
	return nil
}

func (s *session) Read(ctx context.Context, targets []string) ([]protocol.Value, error) {
	nodes, err := parseNodes(targets)
	if err != nil {
		return nil, err
	}
	req := &ua.ReadRequest{TimestampsToReturn: ua.TimestampsToReturnSource}
	for _, node := range nodes {
		req.NodesToRead = append(req.NodesToRead, &ua.ReadValueID{NodeID: node, AttributeID: ua.AttributeIDValue})
	}
	res, err := s.client.Read(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("read OPC UA values: %w", err)
	}
	if len(res.Results) != len(targets) {
		return nil, ua.StatusBadUnknownResponse
	}
	receivedAt := time.Now().UTC()
	values := make([]protocol.Value, len(targets))
	for i, result := range res.Results {
		values[i], err = decodeValue(targets[i], result, receivedAt)
		if err != nil {
			return nil, err
		}
	}
	return values, nil
}

func (s *session) Write(context.Context, string, json.RawMessage) error {
	return fmt.Errorf("OPC UA point writes are not supported")
}

func parseNodes(targets []string) ([]*ua.NodeID, error) {
	if len(targets) == 0 {
		return nil, fmt.Errorf("at least one OPC UA NodeID is required")
	}
	nodes := make([]*ua.NodeID, len(targets))
	for i, target := range targets {
		node, err := ParseNodeID(target)
		if err != nil {
			return nil, fmt.Errorf("OPC UA NodeID %q: %w", target, err)
		}
		nodes[i] = node
	}
	return nodes, nil
}

func ParseNodeID(target string) (*ua.NodeID, error) {
	node, err := ua.ParseNodeID(target)
	if err != nil {
		return nil, err
	}
	if node.String() == "i=0" {
		return nil, fmt.Errorf("null NodeID")
	}
	return node, nil
}

func decodeValue(target string, data *ua.DataValue, receivedAt time.Time) (protocol.Value, error) {
	if data == nil {
		return protocol.Value{}, fmt.Errorf("OPC UA NodeID %q returned an empty DataValue", target)
	}
	var value any
	if data.Value != nil {
		value = data.Value.Value()
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return protocol.Value{}, fmt.Errorf("encode OPC UA value for %q: %w", target, err)
	}
	quality := model.QualityBad
	switch uint32(data.Status) >> 30 {
	case 0:
		quality = model.QualityGood
	case 1:
		quality = model.QualityUncertain
	}
	update := protocol.Value{
		TargetID: target, Value: payload, Quality: string(quality),
		ProtocolStatus: formatStatus(data.Status), ReceivedAt: receivedAt,
	}
	if !data.SourceTimestamp.IsZero() {
		timestamp := data.SourceTimestamp.UTC()
		update.SourceTimestamp = &timestamp
	}
	return update, nil
}

func formatStatus(code ua.StatusCode) string {
	return fmt.Sprintf("0x%08X", uint32(code))
}
