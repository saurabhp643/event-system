package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	App       AppConfig       `yaml:"app"`
	Database  DatabaseConfig  `yaml:"database"`
	Redis     RedisConfig     `yaml:"redis"`
	Auth      AuthConfig      `yaml:"auth"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
	WebSocket WebSocketConfig `yaml:"websocket"`
	Webhooks  WebhooksConfig  `yaml:"webhooks"`
	Logging   LoggingConfig   `yaml:"logging"`
}

// AppConfig represents application settings
type AppConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
	Env  string `yaml:"env"`
}

// DatabaseConfig represents database connection settings
type DatabaseConfig struct {
	Driver          string        `yaml:"driver"`
	Host            string        `yaml:"host"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

// RedisConfig represents Redis connection settings
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"pool_size"`
}

// AuthConfig represents authentication settings
type AuthConfig struct {
	JWTSecret    string        `yaml:"jwt_secret"`
	JWTExpiry    time.Duration `yaml:"jwt_expiry"`
	APIKeyHeader string        `yaml:"api_key_header"`
}

// RateLimitConfig represents rate limiting settings
type RateLimitConfig struct {
	Enabled           bool `yaml:"enabled"`
	RequestsPerMinute int  `yaml:"requests_per_minute"`
	Burst             int  `yaml:"burst"`
}

// WebSocketConfig represents WebSocket settings
type WebSocketConfig struct {
	PingInterval    time.Duration `yaml:"ping_interval"`
	PongTimeout     time.Duration `yaml:"pong_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	ReadBufferSize  int           `yaml:"read_buffer_size"`
	WriteBufferSize int           `yaml:"write_buffer_size"`
}

// WebhooksConfig represents webhook settings
type WebhooksConfig struct {
	Enabled    bool          `yaml:"enabled"`
	MaxRetries int           `yaml:"max_retries"`
	RetryDelay time.Duration `yaml:"retry_delay"`
}

// LoggingConfig represents logging settings
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Override with environment variables if set
	cfg.overrideFromEnv()

	return &cfg, nil
}

// overrideFromEnv overrides configuration with environment variables
func (c *Config) overrideFromEnv() {
	// App Settings
	if host := os.Getenv("APP_HOST"); host != "" {
		c.App.Host = host
	}
	if port := os.Getenv("APP_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			c.App.Port = p
		}
	}
	if mode := os.Getenv("APP_MODE"); mode != "" {
		c.App.Mode = mode
	}
	if env := os.Getenv("APP_ENV"); env != "" {
		c.App.Env = env
	}

	// Database Settings
	if dbPath := os.Getenv("DATABASE_PATH"); dbPath != "" {
		c.Database.Host = dbPath
	}
	if maxOpen := os.Getenv("DATABASE_MAX_OPEN_CONNS"); maxOpen != "" {
		if n, err := strconv.Atoi(maxOpen); err == nil {
			c.Database.MaxOpenConns = n
		}
	}
	if maxIdle := os.Getenv("DATABASE_MAX_IDLE_CONNS"); maxIdle != "" {
		if n, err := strconv.Atoi(maxIdle); err == nil {
			c.Database.MaxIdleConns = n
		}
	}
	if lifetime := os.Getenv("DATABASE_CONN_MAX_LIFETIME"); lifetime != "" {
		if d, err := time.ParseDuration(lifetime); err == nil {
			c.Database.ConnMaxLifetime = d
		}
	}

	// Redis Settings
	if redisHost := os.Getenv("REDIS_HOST"); redisHost != "" {
		c.Redis.Host = redisHost
	}
	if redisPort := os.Getenv("REDIS_PORT"); redisPort != "" {
		if p, err := strconv.Atoi(redisPort); err == nil {
			c.Redis.Port = p
		}
	}
	if redisPassword := os.Getenv("REDIS_PASSWORD"); redisPassword != "" {
		c.Redis.Password = redisPassword
	}
	if redisDB := os.Getenv("REDIS_DB"); redisDB != "" {
		if n, err := strconv.Atoi(redisDB); err == nil {
			c.Redis.DB = n
		}
	}

	// Auth Settings
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		c.Auth.JWTSecret = secret
	}
	if expiry := os.Getenv("JWT_EXPIRY"); expiry != "" {
		if d, err := time.ParseDuration(expiry); err == nil {
			c.Auth.JWTExpiry = d
		}
	}
	if header := os.Getenv("API_KEY_HEADER"); header != "" {
		c.Auth.APIKeyHeader = header
	}

	// Rate Limit Settings
	if enabled := os.Getenv("RATE_LIMIT_ENABLED"); enabled != "" {
		c.RateLimit.Enabled = enabled == "true" || enabled == "1"
	}
	if rpm := os.Getenv("RATE_LIMIT_REQUESTS_PER_MINUTE"); rpm != "" {
		if n, err := strconv.Atoi(rpm); err == nil {
			c.RateLimit.RequestsPerMinute = n
		}
	}
	if burst := os.Getenv("RATE_LIMIT_BURST"); burst != "" {
		if n, err := strconv.Atoi(burst); err == nil {
			c.RateLimit.Burst = n
		}
	}

	// WebSocket Settings
	if ping := os.Getenv("WS_PING_INTERVAL"); ping != "" {
		if d, err := time.ParseDuration(ping); err == nil {
			c.WebSocket.PingInterval = d
		}
	}
	if pong := os.Getenv("WS_PONG_TIMEOUT"); pong != "" {
		if d, err := time.ParseDuration(pong); err == nil {
			c.WebSocket.PongTimeout = d
		}
	}
	if writeTimeout := os.Getenv("WS_WRITE_TIMEOUT"); writeTimeout != "" {
		if d, err := time.ParseDuration(writeTimeout); err == nil {
			c.WebSocket.WriteTimeout = d
		}
	}

	// Webhook Settings
	if enabled := os.Getenv("WEBHOOKS_ENABLED"); enabled != "" {
		c.Webhooks.Enabled = enabled == "true" || enabled == "1"
	}
	if retries := os.Getenv("WEBHOOKS_MAX_RETRIES"); retries != "" {
		if n, err := strconv.Atoi(retries); err == nil {
			c.Webhooks.MaxRetries = n
		}
	}
	if delay := os.Getenv("WEBHOOKS_RETRY_DELAY"); delay != "" {
		if d, err := time.ParseDuration(delay); err == nil {
			c.Webhooks.RetryDelay = d
		}
	}

	// Logging Settings
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		c.Logging.Level = level
	}
	if format := os.Getenv("LOG_FORMAT"); format != "" {
		c.Logging.Format = format
	}
}

// GetRedisAddr returns the Redis address in host:port format
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
