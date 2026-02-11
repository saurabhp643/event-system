package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"event-ingestion-system/internal/database"
	"event-ingestion-system/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	AuthTypeAPIKey = "api_key"
	AuthTypeJWT    = "jwt"
)

// AuthClaims represents the JWT claims
type AuthClaims struct {
	TenantID string `json:"tenant_id"`
	APIKey   string `json:"api_key"`
	jwt.RegisteredClaims
}

// AuthMiddleware provides authentication middleware
type AuthMiddleware struct {
	db           *database.Database
	jwtSecret    []byte
	jwtExpiry    time.Duration
	apiKeyHeader string
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(db *database.Database, jwtSecret string, jwtExpiry time.Duration, apiKeyHeader string) *AuthMiddleware {
	return &AuthMiddleware{
		db:           db,
		jwtSecret:    []byte(jwtSecret),
		jwtExpiry:    jwtExpiry,
		apiKeyHeader: apiKeyHeader,
	}
}

// Authenticate is the main authentication middleware
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try JWT token first
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := m.validateJWT(tokenString)
			if err == nil {
				c.Set("tenant_id", claims.TenantID)
				c.Set("api_key", claims.APIKey)
				c.Set("auth_type", AuthTypeJWT)
				c.Next()
				return
			}
		}

		// Try API key
		apiKey := c.GetHeader(m.apiKeyHeader)
		if apiKey != "" {
			tenant, err := m.db.GetTenantByAPIKey(apiKey)
			if err == nil && tenant.Active {
				c.Set("tenant_id", tenant.ID)
				c.Set("api_key", apiKey)
				c.Set("auth_type", AuthTypeAPIKey)
				c.Set("tenant", tenant)
				c.Next()
				return
			}
		}

		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Invalid or missing authentication credentials",
		})
		c.Abort()
	}
}

// validateJWT validates a JWT token and returns claims
func (m *AuthMiddleware) validateJWT(tokenString string) (*AuthClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*AuthClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token claims")
}

// GenerateJWT generates a JWT token for a tenant
func (m *AuthMiddleware) GenerateJWT(tenant *models.Tenant) (string, error) {
	claims := &AuthClaims{
		TenantID: tenant.ID,
		APIKey:   tenant.APIKey,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.jwtExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "event-ingestion-system",
			Subject:   tenant.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.jwtSecret)
}

// GetTenantFromContext retrieves the tenant from the Gin context
func GetTenantFromContext(c *gin.Context) (*models.Tenant, bool) {
	tenant, exists := c.Get("tenant")
	if !exists {
		return nil, false
	}
	t, ok := tenant.(*models.Tenant)
	return t, ok
}

// GetTenantIDFromContext retrieves the tenant ID from the Gin context
func GetTenantIDFromContext(c *gin.Context) string {
	tenantID, _ := c.Get("tenant_id")
	if id, ok := tenantID.(string); ok {
		return id
	}
	return ""
}

// GetAPIKeyFromContext retrieves the API key from the Gin context
func GetAPIKeyFromContext(c *gin.Context) string {
	apiKey, _ := c.Get("api_key")
	if key, ok := apiKey.(string); ok {
		return key
	}
	return ""
}
