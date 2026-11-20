package model

import "time"

type GatewayState string

const (
	GatewayOnline   GatewayState = "ONLINE"
	GatewayOffline  GatewayState = "OFFLINE"
	GatewayDegraded GatewayState = "DEGRADED"
)

type Gateway struct {
	ID            string            `json:"id"`
	SiteID        string            `json:"siteId"`
	Name          string            `json:"name"`
	Version       string            `json:"version"`
	State         GatewayState      `json:"state"`
	ConfigVersion string            `json:"configVersion"`
	Capabilities  []string          `json:"capabilities,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	LastHeartbeat *time.Time        `json:"lastHeartbeat,omitempty"`
}

type ConnectionState string

const (
	ConnectionDisconnected ConnectionState = "DISCONNECTED"
	ConnectionConnecting   ConnectionState = "CONNECTING"
	ConnectionConnected    ConnectionState = "CONNECTED"
	ConnectionDegraded     ConnectionState = "DEGRADED"
	ConnectionFailed       ConnectionState = "FAILED"
)

type Connection struct {
	ID             string            `json:"id"`
	GatewayID      string            `json:"gatewayId"`
	Protocol       string            `json:"protocol"`
	Endpoint       string            `json:"endpoint"`
	CredentialsRef string            `json:"credentialsRef,omitempty"`
	Options        map[string]string `json:"options,omitempty"`
	State          ConnectionState   `json:"state"`
	LastError      string            `json:"lastError,omitempty"`
}

type SubscriptionMode string

const (
	SubscriptionOnChange SubscriptionMode = "ON_CHANGE"
	SubscriptionPeriodic SubscriptionMode = "PERIODIC"
	SubscriptionEvent    SubscriptionMode = "EVENT"
)

type Subscription struct {
	ID                   string           `json:"id"`
	DeviceID             string           `json:"deviceId"`
	TargetIDs            []string         `json:"targetIds"`
	Mode                 SubscriptionMode `json:"mode"`
	Interval             time.Duration    `json:"interval,omitempty"`
	Deadband             float64          `json:"deadband,omitempty"`
	QueueCapacity        int              `json:"queueCapacity"`
	MaxRatePerSec        float64          `json:"maxRatePerSecond,omitempty"`
	Enabled              bool             `json:"enabled"`
	ConfigurationVersion string           `json:"configurationVersion"`
}
