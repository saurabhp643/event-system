package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Tenant represents a tenant in the multi-tenant system
type Tenant struct {
	ID        string         `gorm:"primaryKey;size:36" json:"id"`
	Name      string         `gorm:"size:255;not null" json:"name"`
	APIKey    string         `gorm:"size:64;uniqueIndex;not null" json:"-"`
	Active    bool           `gorm:"default:true" json:"active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Events   []Event   `gorm:"foreignKey:TenantID" json:"events,omitempty"`
	Webhooks []Webhook `gorm:"foreignKey:TenantID" json:"webhooks,omitempty"`
}

// Event represents an event ingested from a tenant
type Event struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID    string         `gorm:"size:36;index;not null" json:"tenant_id"`
	EventType   string         `gorm:"size:100;index;not null" json:"event_type"`
	Timestamp   time.Time      `gorm:"not null;index" json:"timestamp"`
	Metadata    string         `gorm:"type:text" json:"metadata"` // JSON string
	ProcessedAt *time.Time     `json:"processed_at,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Tenant Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
}

// Webhook represents a webhook endpoint for a tenant (bonus feature)
type Webhook struct {
	ID            uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID      string         `gorm:"size:36;index;not null" json:"tenant_id"`
	URL           string         `gorm:"size:500;not null" json:"url"`
	Secret        string         `gorm:"size:64;not null" json:"-"`
	EventTypes    string         `gorm:"type:text" json:"event_types"` // JSON array
	Active        bool           `gorm:"default:true" json:"active"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
	LastTriggered *time.Time     `json:"last_triggered,omitempty"`
	FailureCount  int            `gorm:"default:0" json:"failure_count"`

	// Relations
	Tenant Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
}

// EventRequest represents the incoming event request
type EventRequest struct {
	TenantID  string          `json:"tenant_id" binding:"required,uuid"`
	EventType string          `json:"event_type" binding:"required,min=1,max=100"`
	Timestamp string          `json:"timestamp" binding:"required"`
	Metadata  json.RawMessage `json:"metadata"`
}

// EventResponse represents an event in the API response
type EventResponse struct {
	ID        uint64          `json:"id"`
	TenantID  string          `json:"tenant_id"`
	EventType string          `json:"event_type"`
	Timestamp time.Time       `json:"timestamp"`
	Metadata  json.RawMessage `json:"metadata"`
	CreatedAt time.Time       `json:"created_at"`
}

// ToEventResponse converts Event to EventResponse
func (e *Event) ToEventResponse() EventResponse {
	var metadata json.RawMessage
	if e.Metadata != "" {
		metadata = json.RawMessage(e.Metadata)
	}
	return EventResponse{
		ID:        uint64(e.ID),
		TenantID:  e.TenantID,
		EventType: e.EventType,
		Timestamp: e.Timestamp,
		Metadata:  metadata,
		CreatedAt: e.CreatedAt,
	}
}

// CreateTenantRequest represents the request to create a tenant
type CreateTenantRequest struct {
	Name string `json:"name" binding:"required,min=1,max=255"`
}

// CreateTenantResponse represents the response after creating a tenant
type CreateTenantResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	APIKey string `json:"api_key"`
}

// AuthToken represents the JWT token payload
type AuthToken struct {
	TenantID string `json:"tenant_id"`
	APIKey   string `json:"api_key"`
	Type     string `json:"type"` // "api_key" or "jwt"
}
