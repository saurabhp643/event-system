package database

import (
	"event-ingestion-system/internal/models"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
	"path/filepath"
	"time"
)

// Database  gorm.DB with configuration
type Database struct {
	DB              *gorm.DB
	Driver          string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// NewDatabase creates a new database connection
func NewDatabase(driver, dsn string, maxOpenConns, maxIdleConns int, connMaxLifetime time.Duration) (*Database, error) {
	var db *gorm.DB
	var err error

	if driver == "postgres" {
		// PostgreSQL connection
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	} else {
		// SQLite connection (default)
		// Create directory if it doesn't exist
		dir := filepath.Dir(dsn)
		if dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("failed to create database directory: %w", err)
			}
		}
		db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	return &Database{
		DB:              db,
		Driver:          driver,
		MaxOpenConns:    maxOpenConns,
		MaxIdleConns:    maxIdleConns,
		ConnMaxLifetime: connMaxLifetime,
	}, nil
}

// BuildDSN builds a PostgreSQL connection string from components
func BuildDSN(host, port, user, password, dbname string) string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
}

// Migrate runs database migrations
func (d *Database) Migrate() error {
	return d.DB.AutoMigrate(
		&models.Tenant{},
		&models.Event{},
		&models.Webhook{},
	)
}

// Close closes the database connection
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// CreateTenant creates a new tenant
func (d *Database) CreateTenant(tenant *models.Tenant) error {
	return d.DB.Create(tenant).Error
}

// GetTenantByID retrieves a tenant by ID
func (d *Database) GetTenantByID(id string) (*models.Tenant, error) {
	var tenant models.Tenant
	err := d.DB.First(&tenant, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

// GetTenantByAPIKey retrieves a tenant by API key
func (d *Database) GetTenantByAPIKey(apiKey string) (*models.Tenant, error) {
	var tenant models.Tenant
	err := d.DB.First(&tenant, "api_key = ?", apiKey).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

// GetTenantByName retrieves a tenant by name
func (d *Database) GetTenantByName(name string) (*models.Tenant, error) {
	var tenant models.Tenant
	err := d.DB.First(&tenant, "name = ?", name).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

// GetAllTenants retrieves all active tenants
func (d *Database) GetAllTenants() ([]models.Tenant, error) {
	var tenants []models.Tenant
	err := d.DB.Where("active = ?", true).Find(&tenants).Error
	return tenants, err
}

// CreateEvent creates a new event
func (d *Database) CreateEvent(event *models.Event) error {
	return d.DB.Create(event).Error
}

// GetEventsByTenant retrieves events for a tenant with pagination
func (d *Database) GetEventsByTenant(tenantID string, limit, offset int) ([]models.Event, error) {
	var events []models.Event
	err := d.DB.Where("tenant_id = ?", tenantID).
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&events).Error
	return events, err
}

// GetEventsByTenantAndType retrieves events for a tenant filtered by event type
func (d *Database) GetEventsByTenantAndType(tenantID, eventType string, limit, offset int) ([]models.Event, error) {
	var events []models.Event
	err := d.DB.Where("tenant_id = ? AND event_type = ?", tenantID, eventType).
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&events).Error
	return events, err
}

// SearchEventsByMetadata searches events by metadata content (basic LIKE search)
func (d *Database) SearchEventsByMetadata(tenantID, query string, limit, offset int) ([]models.Event, error) {
	var events []models.Event
	err := d.DB.Where("tenant_id = ? AND metadata LIKE ?", tenantID, "%"+query+"%").
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&events).Error
	return events, err
}

// GetEventStats retrieves event statistics for a tenant
func (d *Database) GetEventStats(tenantID string) (map[string]int64, error) {
	stats := make(map[string]int64)

	// Total count
	var total int64
	d.DB.Model(&models.Event{}).Where("tenant_id = ?", tenantID).Count(&total)
	stats["total"] = total

	// Count by event type
	var results []struct {
		EventType string
		Count     int64
	}
	d.DB.Model(&models.Event{}).
		Select("event_type, COUNT(*) as count").
		Where("tenant_id = ?", tenantID).
		Group("event_type").
		Find(&results)

	for _, r := range results {
		stats[r.EventType] = r.Count
	}

	return stats, nil
}

// CreateWebhook creates a new webhook
func (d *Database) CreateWebhook(webhook *models.Webhook) error {
	return d.DB.Create(webhook).Error
}

// GetWebhooksByTenant retrieves webhooks for a tenant
func (d *Database) GetWebhooksByTenant(tenantID string) ([]models.Webhook, error) {
	var webhooks []models.Webhook
	err := d.DB.Where("tenant_id = ? AND active = ?", tenantID, true).Find(&webhooks).Error
	return webhooks, err
}
