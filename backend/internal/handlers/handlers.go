package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"event-ingestion-system/internal/auth"
	"event-ingestion-system/internal/database"
	"event-ingestion-system/internal/errors"
	"event-ingestion-system/internal/models"
	"event-ingestion-system/internal/websocket"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	db   *database.Database
	hub  *websocket.Hub
	auth *auth.AuthMiddleware
}

// NewHandler creates a new handler
func NewHandler(db *database.Database, hub *websocket.Hub, authMiddleware *auth.AuthMiddleware) *Handler {
	return &Handler{
		db:   db,
		hub:  hub,
		auth: authMiddleware,
	}
}

// GetDB returns the database instance
func (h *Handler) GetDB() *database.Database {
	return h.db
}

// GetHub returns the WebSocket hub
func (h *Handler) GetHub() *websocket.Hub {
	return h.hub
}

// HealthCheck returns the health status of the API
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "1.0.0",
	})
}

// CreateTenant creates a new tenant with validation
func (h *Handler) CreateTenant(c *gin.Context) {
	var req models.CreateTenantRequest

	// Parse and validate JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.ErrInvalidRequest(err.Error()).Response())
		return
	}

	// Validate tenant name
	if err := validateTenantName(req.Name); err != nil {
		c.JSON(http.StatusBadRequest, errors.ErrInvalidRequest(err.Error()).Response())
		return
	}

	// Check if tenant with same name exists
	existing, err := h.db.GetTenantByName(req.Name)
	if err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, errors.ErrInternal("Failed to check existing tenant", err).Response())
		return
	}
	if existing != nil {
		c.JSON(http.StatusConflict, errors.ErrTenantExists(req.Name).Response())
		return
	}

	tenant := &models.Tenant{
		ID:     uuid.New().String(),
		Name:   req.Name,
		APIKey: uuid.New().String(),
		Active: true,
	}

	if err := h.db.CreateTenant(tenant); err != nil {
		c.JSON(http.StatusInternalServerError, errors.ErrDB("create tenant", err).Response())
		return
	}

	// Generate JWT token
	token, _ := h.auth.GenerateJWT(tenant)

	c.JSON(http.StatusCreated, gin.H{
		"id":         tenant.ID,
		"name":       tenant.Name,
		"api_key":    tenant.APIKey,
		"token":      token,
		"active":     tenant.Active,
		"created_at": tenant.CreatedAt.Format(time.RFC3339),
	})
}

// GetTenants returns all tenants (without API keys for security)
func (h *Handler) GetTenants(c *gin.Context) {
	tenants, err := h.db.GetAllTenants()
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.ErrDB("get tenants", err).Response())
		return
	}

	// Hide API keys in response for security
	type TenantResponse struct {
		ID        string    `json:"id"`
		Name      string    `json:"name"`
		Active    bool      `json:"active"`
		CreatedAt time.Time `json:"created_at"`
	}

	response := make([]TenantResponse, 0, len(tenants))
	for _, t := range tenants {
		response = append(response, TenantResponse{
			ID:        t.ID,
			Name:      t.Name,
			Active:    t.Active,
			CreatedAt: t.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"tenants": response})
}

// GetTenantsWithKeys returns all tenants WITH API keys (for frontend use only)
func (h *Handler) GetTenantsWithKeys(c *gin.Context) {
	tenants, err := h.db.GetAllTenants()
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.ErrDB("get tenants", err).Response())
		return
	}

	response := make([]gin.H, 0, len(tenants))
	for _, t := range tenants {
		response = append(response, gin.H{
			"id":         t.ID,
			"name":       t.Name,
			"api_key":    t.APIKey,
			"active":     t.Active,
			"created_at": t.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"tenants": response})
}

// GetTenant returns a specific tenant
func (h *Handler) GetTenant(c *gin.Context) {
	tenantID := c.Param("id")

	// Validate UUID format
	if _, err := uuid.Parse(tenantID); err != nil {
		c.JSON(http.StatusBadRequest, errors.ErrBadTenantID("Invalid UUID format").Response())
		return
	}

	tenant, err := h.db.GetTenantByID(tenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, errors.ErrTenantNotFound(tenantID).Response())
			return
		}
		c.JSON(http.StatusInternalServerError, errors.ErrDB("get tenant", err).Response())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         tenant.ID,
		"name":       tenant.Name,
		"active":     tenant.Active,
		"api_key":    tenant.APIKey,
		"created_at": tenant.CreatedAt.Format(time.RFC3339),
	})
}

// IngestEvent ingests a new event with comprehensive validation
func (h *Handler) IngestEvent(c *gin.Context) {
	var req models.EventRequest

	// Parse and validate JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.ErrInvalidRequest(err.Error()).Response())
		return
	}

	// Validate tenant ID
	if _, err := uuid.Parse(req.TenantID); err != nil {
		c.JSON(http.StatusBadRequest, errors.ErrBadTenantID("Invalid tenant ID format").Response())
		return
	}

	// Check tenant exists and is active
	tenant, err := h.db.GetTenantByID(req.TenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, errors.ErrTenantNotFound(req.TenantID).Response())
			return
		}
		c.JSON(http.StatusInternalServerError, errors.ErrDB("verify tenant", err).Response())
		return
	}
	if !tenant.Active {
		c.JSON(http.StatusForbidden, errors.ErrUnauthorized("Tenant is inactive").Response())
		return
	}

	// Validate event type
	if err := validateEventType(req.EventType); err != nil {
		c.JSON(http.StatusBadRequest, errors.ErrBadEventType(err.Error()).Response())
		return
	}

	// Parse timestamp - support multiple formats
	timestamp, err := parseTimestamp(req.Timestamp)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.ErrBadTimestamp("Timestamp must be in ISO8601 format (e.g., 2026-02-10T19:07:41Z or 2026-02-10T19:07:41.701Z)").Response())
		return
	}

	// Validate metadata is valid JSON
	if req.Metadata != nil {
		if _, err := json.Marshal(req.Metadata); err != nil {
			c.JSON(http.StatusBadRequest, errors.ErrBadMetadata("Metadata must be a valid JSON object").Response())
			return
		}
	}

	metadata, _ := json.Marshal(req.Metadata)
	event := &models.Event{
		TenantID:  req.TenantID,
		EventType: req.EventType,
		Timestamp: timestamp,
		Metadata:  string(metadata),
	}

	if err := h.db.CreateEvent(event); err != nil {
		c.JSON(http.StatusInternalServerError, errors.ErrDB("create event", err).Response())
		return
	}

	// Broadcast to WebSocket clients (non-blocking)
	go h.hub.BroadcastToTenant(req.TenantID, event)

	c.JSON(http.StatusCreated, gin.H{
		"id":         event.ID,
		"tenant_id":  event.TenantID,
		"event_type": event.EventType,
		"timestamp":  event.Timestamp.Format(time.RFC3339),
	})
}

// GetEvents returns events for a tenant with filtering and pagination
func (h *Handler) GetEvents(c *gin.Context) {
	tenantID := c.GetString("tenant_id")

	// Parse query parameters with validation
	limit := 50
	offset := 0

	if l := c.Query("limit"); l != "" {
		parsed, err := strconv.Atoi(l)
		if err != nil || parsed < 0 {
			c.JSON(http.StatusBadRequest, errors.ErrInvalidRequest("Invalid limit parameter").Response())
			return
		}
		if parsed > 100 {
			parsed = 100 // Cap at 100
		}
		limit = parsed
	}

	if o := c.Query("offset"); o != "" {
		parsed, err := strconv.Atoi(o)
		if err != nil || parsed < 0 {
			c.JSON(http.StatusBadRequest, errors.ErrInvalidRequest("Invalid offset parameter").Response())
			return
		}
		offset = parsed
	}

	eventType := c.Query("event_type")
	search := c.Query("search")

	var events []models.Event
	var fetchErr error

	if eventType != "" {
		// Validate event type
		if err := validateEventType(eventType); err != nil {
			c.JSON(http.StatusBadRequest, errors.ErrBadEventType(err.Error()).Response())
			return
		}
		events, fetchErr = h.db.GetEventsByTenantAndType(tenantID, eventType, limit, offset)
	} else if search != "" {
		events, fetchErr = h.db.SearchEventsByMetadata(tenantID, search, limit, offset)
	} else {
		events, fetchErr = h.db.GetEventsByTenant(tenantID, limit, offset)
	}

	if fetchErr != nil {
		c.JSON(http.StatusInternalServerError, errors.ErrDB("get events", fetchErr).Response())
		return
	}

	response := make([]models.EventResponse, 0, len(events))
	for _, e := range events {
		response = append(response, e.ToEventResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"events": response,
		"limit":  limit,
		"offset": offset,
	})
}

// GetEventStats returns event statistics for a tenant
func (h *Handler) GetEventStats(c *gin.Context) {
	tenantID := c.GetString("tenant_id")

	stats, err := h.db.GetEventStats(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.ErrDB("get event stats", err).Response())
		return
	}

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

// GetAuthToken generates a JWT token for a tenant
func (h *Handler) GetAuthToken(c *gin.Context) {
	tenantID := c.Param("id")

	// Validate UUID format
	if _, err := uuid.Parse(tenantID); err != nil {
		c.JSON(http.StatusBadRequest, errors.ErrBadTenantID("Invalid UUID format").Response())
		return
	}

	tenant, err := h.db.GetTenantByID(tenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, errors.ErrTenantNotFound(tenantID).Response())
			return
		}
		c.JSON(http.StatusInternalServerError, errors.ErrDB("get tenant", err).Response())
		return
	}

	token, err := h.auth.GenerateJWT(tenant)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errors.ErrInternal("Failed to generate token", err).Response())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":      token,
		"token_type": "Bearer",
		"expires_in": 86400,
	})
}

// Helper functions for validation

// validateTenantName validates the tenant name
func validateTenantName(name string) error {
	if len(name) < 3 {
		return &ValidationError{Field: "name", Message: "must be at least 3 characters"}
	}
	if len(name) > 50 {
		return &ValidationError{Field: "name", Message: "must be at most 50 characters"}
	}
	return nil
}

// validateEventType validates the event type
func validateEventType(eventType string) error {
	if len(eventType) < 1 {
		return &ValidationError{Field: "event_type", Message: "cannot be empty"}
	}
	if len(eventType) > 100 {
		return &ValidationError{Field: "event_type", Message: "must be at most 100 characters"}
	}
	// Allow alphanumeric characters, underscores, hyphens, and dots
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_.-]+$`, eventType)
	if !matched {
		return &ValidationError{Field: "event_type", Message: "can only contain alphanumeric characters, underscores, hyphens, and dots"}
	}
	return nil
}

// parseTimestamp parses timestamp in various ISO8601 formats
func parseTimestamp(ts string) (time.Time, error) {
	// Try multiple formats
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05.000000Z",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, ts); err == nil {
			return t, nil
		}
	}

	return time.Time{}, &ValidationError{Field: "timestamp", Message: "invalid format"}
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}
