package errors

import (
	"net/http"
	"strconv"
)

// ErrorCode represents a structured error code
type ErrorCode string

const (
	// Validation errors (400)
	CodeInvalidRequest   ErrorCode = "invalid_request"
	CodeInvalidTenantID  ErrorCode = "invalid_tenant_id"
	CodeInvalidEventType ErrorCode = "invalid_event_type"
	CodeInvalidTimestamp ErrorCode = "invalid_timestamp"
	CodeInvalidMetadata  ErrorCode = "invalid_metadata"

	// Authentication errors (401)
	CodeUnauthorized  ErrorCode = "unauthorized"
	CodeInvalidAPIKey ErrorCode = "invalid_api_key"
	CodeExpiredToken  ErrorCode = "expired_token"
	CodeMissingAuth   ErrorCode = "missing_authentication"

	// Not found errors (404)
	CodeTenantNotFound ErrorCode = "tenant_not_found"
	CodeEventNotFound  ErrorCode = "event_not_found"

	// Conflict errors (409)
	CodeTenantExists ErrorCode = "tenant_exists"

	// Rate limit errors (429)
	CodeRateLimitExceeded ErrorCode = "rate_limit_exceeded"

	// Server errors (500)
	CodeInternalError  ErrorCode = "internal_error"
	CodeDatabaseError  ErrorCode = "database_error"
	CodeWebSocketError ErrorCode = "websocket_error"
)

// AppError represents a structured application error
type AppError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	StatusCode int       `json:"-"`
	Internal   error     `json:"-"`
}

// NewAppError creates a new application error
func NewAppError(code ErrorCode, message, details string, statusCode int, internal error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Details:    details,
		StatusCode: statusCode,
		Internal:   internal,
	}
}

// Validation errors
func ErrInvalidRequest(details string) *AppError {
	return NewAppError(CodeInvalidRequest, "Invalid request", details, http.StatusBadRequest, nil)
}

func ErrBadTenantID(details string) *AppError {
	return NewAppError(CodeInvalidTenantID, "Invalid tenant ID", details, http.StatusBadRequest, nil)
}

func ErrBadEventType(details string) *AppError {
	return NewAppError(CodeInvalidEventType, "Invalid event type", details, http.StatusBadRequest, nil)
}

func ErrBadTimestamp(details string) *AppError {
	return NewAppError(CodeInvalidTimestamp, "Invalid timestamp format", details, http.StatusBadRequest, nil)
}

func ErrBadMetadata(details string) *AppError {
	return NewAppError(CodeInvalidMetadata, "Invalid metadata", details, http.StatusBadRequest, nil)
}

// Authentication errors
func ErrUnauthorized(details string) *AppError {
	return NewAppError(CodeUnauthorized, "Unauthorized", details, http.StatusUnauthorized, nil)
}

func ErrBadAPIKey() *AppError {
	return NewAppError(CodeInvalidAPIKey, "Invalid API key", "The provided API key is not valid", http.StatusUnauthorized, nil)
}

func ErrTokenExpired() *AppError {
	return NewAppError(CodeExpiredToken, "Token expired", "The authentication token has expired", http.StatusUnauthorized, nil)
}

func ErrNoAuth() *AppError {
	return NewAppError(CodeMissingAuth, "Missing authentication", "No authentication credentials provided", http.StatusUnauthorized, nil)
}

// Not found errors
func ErrTenantNotFound(tenantID string) *AppError {
	return NewAppError(CodeTenantNotFound, "Tenant not found", "Tenant with ID '"+tenantID+"' was not found", http.StatusNotFound, nil)
}

func ErrEventNotFound(eventID int) *AppError {
	return NewAppError(CodeEventNotFound, "Event not found", "Event with ID '"+strconv.Itoa(eventID)+"' was not found", http.StatusNotFound, nil)
}

// Conflict errors
func ErrTenantExists(name string) *AppError {
	return NewAppError(CodeTenantExists, "Tenant already exists", "A tenant with name '"+name+"' already exists", http.StatusConflict, nil)
}

// Rate limit errors
func ErrRateLimit() *AppError {
	return NewAppError(CodeRateLimitExceeded, "Rate limit exceeded", "Too many requests. Please try again later.", http.StatusTooManyRequests, nil)
}

// Server errors
func ErrInternal(details string, internal error) *AppError {
	return NewAppError(CodeInternalError, "Internal server error", details, http.StatusInternalServerError, internal)
}

func ErrDB(operation string, internal error) *AppError {
	return NewAppError(CodeDatabaseError, "Database operation failed", "Failed to "+operation, http.StatusInternalServerError, internal)
}

func ErrWS(internal error) *AppError {
	return NewAppError(CodeWebSocketError, "WebSocket connection failed", "Unable to establish WebSocket connection", http.StatusInternalServerError, internal)
}

// Error returns the error message
func (e *AppError) Error() string {
	if e.Internal != nil {
		return e.Message + ": " + e.Internal.Error()
	}
	return e.Message
}

// Unwrap returns the internal error
func (e *AppError) Unwrap() error {
	return e.Internal
}

// Response returns a structured error response
func (e *AppError) Response() map[string]interface{} {
	response := map[string]interface{}{
		"error": map[string]interface{}{
			"code":    e.Code,
			"message": e.Message,
		},
	}
	if e.Details != "" {
		response["error"].(map[string]interface{})["details"] = e.Details
	}
	return response
}

// Is checks if the error is of a specific type
func (e *AppError) Is(code ErrorCode) bool {
	return e.Code == code
}
