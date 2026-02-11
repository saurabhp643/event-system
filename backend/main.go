package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"event-ingestion-system/internal/auth"
	"event-ingestion-system/internal/config"
	"event-ingestion-system/internal/database"
	"event-ingestion-system/internal/handlers"
	"event-ingestion-system/internal/middleware"
	"event-ingestion-system/internal/websocket"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Set Gin mode
	gin.SetMode(cfg.App.Mode)

	// Initialize database
	// Build DSN based on driver
	var dsn string
	switch cfg.Database.Driver {
	case "postgres":
		// PostgreSQL DSN format
		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
			cfg.Database.Host,
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
			os.Getenv("DB_PORT"))
	default:
		// SQLite DSN is just the file path
		dsn = cfg.Database.Host
	}

	db, err := database.NewDatabase(
		cfg.Database.Driver,
		dsn,
		cfg.Database.MaxOpenConns,
		cfg.Database.MaxIdleConns,
		cfg.Database.ConnMaxLifetime,
	)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.Migrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize WebSocket hub
	wsCfg := &config.WebSocketConfig{
		PingInterval:    cfg.WebSocket.PingInterval,
		PongTimeout:     cfg.WebSocket.PongTimeout,
		WriteTimeout:    cfg.WebSocket.WriteTimeout,
		ReadBufferSize:  cfg.WebSocket.ReadBufferSize,
		WriteBufferSize: cfg.WebSocket.WriteBufferSize,
	}
	hub := websocket.NewHub(wsCfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go hub.Run(ctx)

	// Initialize auth middleware
	authMiddleware := auth.NewAuthMiddleware(
		db,
		cfg.Auth.JWTSecret,
		cfg.Auth.JWTExpiry,
		cfg.Auth.APIKeyHeader,
	)

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimit.RequestsPerMinute)

	// Initialize handlers
	handler := handlers.NewHandler(db, hub, authMiddleware)

	// Setup router
	router := setupRouter(handler, authMiddleware, rateLimiter, cfg, db)

	// Create server
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port),
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on %s:%d", cfg.App.Host, cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func setupRouter(handler *handlers.Handler, authMiddleware *auth.AuthMiddleware, rateLimiter *middleware.RateLimiter, cfg *config.Config, db *database.Database) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(corsMiddleware())

	// Debug endpoint to show all routes
	router.GET("/debug/routes", func(c *gin.Context) {
		routes := router.Routes()
		c.JSON(200, gin.H{"routes": routes})
	})

	// Health check (no auth required)
	router.GET("/health", handler.HealthCheck)

	// API v1 - Public routes (no auth required)
	router.POST("/api/v1/tenants", handler.CreateTenant)
	router.GET("/api/v1/tenants", handler.GetTenants)
	router.GET("/api/v1/tenants-with-keys", handler.GetTenantsWithKeys)

	// API v1 - Protected routes (auth required)
	protected := router.Group("/api/v1")
	protected.Use(authMiddleware.Authenticate())
	protected.Use(middleware.RateLimitMiddleware(rateLimiter, cfg.RateLimit.Enabled))
	{
		// Tenants
		protected.GET("/tenants/:id", handler.GetTenant)
		protected.GET("/tenants/:id/token", handler.GetAuthToken)

		// Events
		protected.POST("/events", handler.IngestEvent)
		protected.GET("/events", handler.GetEvents)
		protected.GET("/events/stats", handler.GetEventStats)
	}

	// WebSocket endpoint
	router.GET("/api/v1/ws", func(c *gin.Context) {
		// Try to authenticate from query param first
		apiKey := c.Query("api_key")
		if apiKey != "" {
			tenant, err := db.GetTenantByAPIKey(apiKey)
			if err == nil && tenant.Active {
				c.Set("tenant_id", tenant.ID)
				c.Set("api_key", apiKey)
				c.Set("auth_type", "api_key")
				hub := handler.GetHub()
				hub.HandleWebSocket(c)
				return
			}
		}
		// Fall back to normal auth
		authMiddleware.Authenticate()(c)
		hub := handler.GetHub()
		hub.HandleWebSocket(c)
	})

	return router
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-API-Key")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
