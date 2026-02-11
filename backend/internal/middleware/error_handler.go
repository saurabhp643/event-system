package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"event-ingestion-system/internal/errors"

	"github.com/gin-gonic/gin"
)

// ErrorHandler is a middleware that handles panics and structured errors
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic with stack trace
				log.Printf("[PANIC] %v\n%s", err, debug.Stack())

				// Check if it's an AppError
				if appErr, ok := err.(*errors.AppError); ok {
					c.JSON(appErr.StatusCode, appErr.Response())
					return
				}

				// Generic panic response
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{
						"code":    errors.CodeInternalError,
						"message": "An unexpected error occurred",
					},
				})
			}
		}()

		c.Next()
	}
}

// RequestLogger logs all requests with timing and status
func RequestLogger() gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/health", "/debug/routes"},
	})
}

// SecurityHeaders adds security headers to responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
		c.Header("Pragma", "no-cache")

		c.Next()
	}
}

// RequestTimeout adds request timeout handling
func RequestTimeout(timeoutSeconds int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set timeout on context
		// Note: For Gin, this is handled at server level
		// This middleware can be used for additional timeout logic if needed

		c.Next()
	}
}

// RequestID adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Request ID is already handled by Gin in production setups
		// This can be extended to add custom request ID logic

		c.Next()
	}
}
