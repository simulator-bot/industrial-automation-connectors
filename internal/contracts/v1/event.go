package v1

import (
	"encoding/json"
	"time"
)

type PointEvent struct {
	SchemaVersion   string          `json:"schemaVersion"`
	EventID         string          `json:"eventId"`
	GatewayID       string          `json:"gatewayId"`
	ConnectionID    string          `json:"connectionId"`
	DeviceID        string          `json:"deviceId"`
	PointID         string          `json:"pointId"`
	Value           json.RawMessage `json:"value"`
	DataType        string          `json:"dataType"`
	Quality         string          `json:"quality"`
	SourceTimestamp *time.Time      `json:"sourceTimestamp,omitempty"`
	IngestTimestamp time.Time       `json:"ingestTimestamp"`
	Sequence        uint64          `json:"sequence,omitempty"`
}
