package v1

import "encoding/json"

type OPCUAPoint struct {
	ID       string `json:"id"`
	NodeID   string `json:"nodeId"`
	DataType string `json:"dataType"`
}

type OPCUAMethod struct {
	ObjectNodeID string `json:"objectNodeId"`
	MethodNodeID string `json:"methodNodeId"`
}

type OPCUAArgument struct {
	Type  string          `json:"type"`
	Value json.RawMessage `json:"value"`
}

type OPCUAMethodTest struct {
	OPCUAMethod
	Arguments []OPCUAArgument `json:"arguments"`
}

type OPCUATestRequest struct {
	Endpoint       string           `json:"endpoint"`
	DeviceID       string           `json:"deviceId"`
	IntervalMillis int              `json:"intervalMillis"`
	Deadband       float64          `json:"deadband"`
	Points         []OPCUAPoint     `json:"points,omitempty"`
	Method         *OPCUAMethodTest `json:"method,omitempty"`
}

type OPCUATestPoint struct {
	PointID       string      `json:"pointId"`
	Read          *PointEvent `json:"read,omitempty"`
	Latest        *PointEvent `json:"latest,omitempty"`
	Notifications int         `json:"notifications"`
}

type OPCUATestResult struct {
	Connected          bool             `json:"connected"`
	ReadCompleted      bool             `json:"readCompleted"`
	SubscriptionOpened bool             `json:"subscriptionOpened"`
	Points             []OPCUATestPoint `json:"points,omitempty"`
	MethodResult       json.RawMessage  `json:"methodResult,omitempty"`
	Error              string           `json:"error,omitempty"`
}
