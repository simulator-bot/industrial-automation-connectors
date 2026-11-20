package model

import "time"

type Device struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Vendor       string            `json:"vendor,omitempty"`
	Model        string            `json:"model,omitempty"`
	SerialNumber string            `json:"serialNumber,omitempty"`
	ConnectionID string            `json:"connectionId"`
	Tags         map[string]string `json:"tags,omitempty"`
	CreatedAt    time.Time         `json:"createdAt"`
	UpdatedAt    time.Time         `json:"updatedAt"`
}

type DataType string

const (
	DataTypeBool   DataType = "bool"
	DataTypeInt    DataType = "int"
	DataTypeUint   DataType = "uint"
	DataTypeFloat  DataType = "float"
	DataTypeString DataType = "string"
	DataTypeBytes  DataType = "bytes"
	DataTypeEnum   DataType = "enum"
	DataTypeTime   DataType = "time"
	DataTypeStruct DataType = "struct"
)

type Access struct {
	Readable     bool `json:"readable"`
	Writable     bool `json:"writable"`
	Subscribable bool `json:"subscribable"`
}

type Point struct {
	ID           string            `json:"id"`
	DeviceID     string            `json:"deviceId"`
	Name         string            `json:"name"`
	Path         string            `json:"path"`
	DataType     DataType          `json:"dataType"`
	Unit         string            `json:"unit,omitempty"`
	Access       Access            `json:"access"`
	ProtocolData map[string]string `json:"protocolData,omitempty"`
}

type Function struct {
	ID           string         `json:"id"`
	DeviceID     string         `json:"deviceId"`
	Name         string         `json:"name"`
	Description  string         `json:"description,omitempty"`
	RiskLevel    string         `json:"riskLevel"`
	InputSchema  map[string]any `json:"inputSchema,omitempty"`
	OutputSchema map[string]any `json:"outputSchema,omitempty"`
	Timeout      time.Duration  `json:"timeout"`
}

type Quality string

const (
	QualityGood      Quality = "GOOD"
	QualityUncertain Quality = "UNCERTAIN"
	QualityBad       Quality = "BAD"
)
