package model

type EquipmentCatalog struct {
	Revision  uint64           `json:"revision"`
	Models    []DeviceModel    `json:"models"`
	Points    []ModelPoint     `json:"points"`
	Plans     []CollectionPlan `json:"plans"`
	Instances []DeviceInstance `json:"instances"`
}

type DeviceModel struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Protocol    string        `json:"protocol"`
	Description string        `json:"description"`
	Methods     []ModelMethod `json:"methods"`
}

type ModelMethod struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	ObjectNodeID string            `json:"objectNodeId"`
	MethodNodeID string            `json:"methodNodeId"`
	Parameters   []MethodParameter `json:"parameters"`
}

type MethodParameter struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type ModelPoint struct {
	ID       string   `json:"id"`
	ModelID  string   `json:"modelId"`
	Name     string   `json:"name"`
	NodeID   string   `json:"nodeId"`
	DataType DataType `json:"dataType"`
	Unit     string   `json:"unit"`
}

type CollectionPlan struct {
	ID             string           `json:"id"`
	ModelID        string           `json:"modelId"`
	Name           string           `json:"name"`
	Mode           SubscriptionMode `json:"mode"`
	PointIDs       []string         `json:"pointIds"`
	IntervalMillis int              `json:"intervalMillis"`
	Deadband       float64          `json:"deadband"`
}

type DeviceInstance struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ModelID  string `json:"modelId"`
	PlanID   string `json:"planId"`
	Endpoint string `json:"endpoint"`
}
