package v1

import "time"

type CommandOperation string

const (
	CommandWrite  CommandOperation = "WRITE"
	CommandInvoke CommandOperation = "INVOKE"
)

type CommandStatus string

const (
	CommandReceived   CommandStatus = "RECEIVED"
	CommandValidated  CommandStatus = "VALIDATED"
	CommandDispatched CommandStatus = "DISPATCHED"
	CommandSucceeded  CommandStatus = "SUCCEEDED"
	CommandFailed     CommandStatus = "FAILED"
	CommandTimedOut   CommandStatus = "TIMED_OUT"
	CommandCancelled  CommandStatus = "CANCELLED"
	CommandUnknown    CommandStatus = "UNKNOWN"
)

// AuthorizedCommand is issued by Process only after policy checks. Gateways
// must validate ownership/configuration fencing before protocol execution.
type AuthorizedCommand struct {
	SchemaVersion  string           `json:"schemaVersion"`
	CommandID      string           `json:"commandId"`
	IdempotencyKey string           `json:"idempotencyKey,omitempty"`
	GatewayID      string           `json:"gatewayId"`
	DeviceID       string           `json:"deviceId"`
	TargetID       string           `json:"targetId"`
	Operation      CommandOperation `json:"operation"`
	Parameters     map[string]any   `json:"parameters,omitempty"`
	Initiator      string           `json:"initiator"`
	PolicyVersion  string           `json:"policyVersion"`
	ConfigVersion  string           `json:"configVersion"`
	FencingToken   uint64           `json:"fencingToken"`
	ExpiresAt      time.Time        `json:"expiresAt"`
}

type CommandResult struct {
	CommandID string         `json:"commandId"`
	Status    CommandStatus  `json:"status"`
	Message   string         `json:"message,omitempty"`
	Result    map[string]any `json:"result,omitempty"`
	UpdatedAt time.Time      `json:"updatedAt"`
}
