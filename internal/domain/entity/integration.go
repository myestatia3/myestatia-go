package entity

import "time"

type IntegrationType string

const (
	IntegrationTypeEmail   IntegrationType = "email"
	IntegrationTypeResales IntegrationType = "resales"
)

type IntegrationStatus string

const (
	IntegrationStatusConnected    IntegrationStatus = "connected"
	IntegrationStatusDisconnected IntegrationStatus = "disconnected"
	IntegrationStatusError        IntegrationStatus = "error"
)

type IntegrationConfig struct {
	ID        string            `json:"id"`
	Type      IntegrationType   `json:"type"`
	Name      string            `json:"name"`
	Status    IntegrationStatus `json:"status"`
	Settings  map[string]string `json:"settings"` // Stores API Keys, URLs, etc. securely
	LastSync  *time.Time        `json:"last_sync,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

func (c *IntegrationConfig) GetResalesAPIKey() string {
	if c.Settings == nil {
		return ""
	}
	return c.Settings["api_key"]
}

func (c *IntegrationConfig) GetResalesAgencyID() string {
	if c.Settings == nil {
		return ""
	}
	return c.Settings["agency_id"]
}
