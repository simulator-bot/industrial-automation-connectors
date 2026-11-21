package protocol

import (
	"context"
	"encoding/json"
	"time"
)

type Descriptor struct {
	Protocol string            `json:"protocol"`
	Version  string            `json:"version"`
	Vendor   string            `json:"vendor,omitempty"`
	Features []string          `json:"features"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type ConnectionConfig struct {
	ConnectionID string            `json:"connectionId"`
	Endpoint     string            `json:"endpoint"`
	Credentials  string            `json:"credentialsRef,omitempty"`
	Options      map[string]string `json:"options,omitempty"`
}

type Session interface {
	Health(context.Context) error
	Read(context.Context, []string) ([]Value, error)
	// Subscribe creates a stream that closes on context cancellation, a
	// terminal update error, or session closure.
	Subscribe(context.Context, Subscription) (<-chan Update, error)
	Write(context.Context, string, json.RawMessage) error
	Invoke(context.Context, string, json.RawMessage) (json.RawMessage, error)
	Close() error
}

type Value struct {
	TargetID        string
	Value           json.RawMessage
	Quality         string
	ProtocolStatus  string
	SourceTimestamp *time.Time
	ReceivedAt      time.Time
	Sequence        uint64
}

type Update struct {
	Value Value
	Err   error
}

type Subscription struct {
	ID       string
	Targets  []string
	Mode     string
	Interval time.Duration
	Deadband float64
}

// Adapter isolates a protocol stack. Gateway owns every returned session.
type Adapter interface {
	Descriptor() Descriptor
	Connect(context.Context, ConnectionConfig) (Session, error)
}
