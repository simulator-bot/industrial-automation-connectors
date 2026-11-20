package rules

// Definition is the transport-neutral boundary for the future CEL + bounded
// action-graph runtime. It intentionally contains no user-supplied Go code.
type Definition struct {
	ID           string          `json:"id"`
	Version      uint64          `json:"version"`
	State        string          `json:"state"`
	EventSchema  string          `json:"eventSchema"`
	ConditionCEL string          `json:"conditionCel"`
	Actions      []Action        `json:"actions"`
	Limits       ExecutionLimits `json:"limits"`
}

type Action struct {
	Kind       string         `json:"kind"`
	Parameters map[string]any `json:"parameters,omitempty"`
}

type ExecutionLimits struct {
	MaxDurationMillis int `json:"maxDurationMillis"`
	MaxActions        int `json:"maxActions"`
	MaxWindowEvents   int `json:"maxWindowEvents"`
}
