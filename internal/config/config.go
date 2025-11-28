package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Server      ServerConfig      `json:"server"`
	Database    DatabaseConfig    `json:"database"`
	JWT         JWTConfig         `json:"jwt"`
	Permissions PermissionsConfig `json:"permissions"`
	Security    SecurityConfig    `json:"security"`
	Logging     LoggingConfig     `json:"logging"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port            string        `json:"port"`
	Host            string        `json:"host"`
	ReadTimeout     time.Duration `json:"read_timeout"`
	WriteTimeout    time.Duration `json:"write_timeout"`
	IdleTimeout     time.Duration `json:"idle_timeout"`
	ShutdownTimeout time.Duration `json:"shutdown_timeout"`
	Environment     string        `json:"environment"`
	Version         string        `json:"version"`
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	User            string        `json:"user"`
	Password        string        `json:"password"`
	Name            string        `json:"name"`
	SSLMode         string        `json:"ssl_mode"`
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time"`
	MigrationPath   string        `json:"migration_path"`
}

// JWTConfig holds JWT-related configuration
type JWTConfig struct {
	Secret            string        `json:"secret"`
	Expiration        time.Duration `json:"expiration"`
	RefreshExpiration time.Duration `json:"refresh_expiration"`
}

// PermissionsConfig holds permission masks for roles
type PermissionsConfig struct {
	Admin    int `json:"admin"`
	Owner    int `json:"owner"`
	Manager  int `json:"manager"`
	Mechanic int `json:"mechanic"`
	Driver   int `json:"driver"`
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	BCryptCost         int      `json:"bcrypt_cost"`
	RateLimitRPS       int      `json:"rate_limit_rps"`
	RateLimitBurst     int      `json:"rate_limit_burst"`
	CORSAllowedOrigins []string `json:"cors_allowed_origins"`
	CORSAllowedMethods []string `json:"cors_allowed_methods"`
	CORSAllowedHeaders []string `json:"cors_allowed_headers"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"`
	Output     string `json:"output"`
	MaxSize    int    `json:"max_size"`
	MaxBackups int    `json:"max_backups"`
	MaxAge     int    `json:"max_age"`
	Compress   bool   `json:"compress"`
}

// Load loads configuration from environment variables and .env file
func Load() (*Config, error) {
	// Try to load .env file (ignore error if file doesn't exist)
	_ = godotenv.Load()

	config := &Config{
		Server: ServerConfig{
			Port:            getEnv("SERVER_PORT", "8080"),
			Host:            getEnv("SERVER_HOST", "0.0.0.0"),
			ReadTimeout:     getDurationEnv("SERVER_READ_TIMEOUT", "30s"),
			WriteTimeout:    getDurationEnv("SERVER_WRITE_TIMEOUT", "30s"),
			IdleTimeout:     getDurationEnv("SERVER_IDLE_TIMEOUT", "60s"),
			ShutdownTimeout: getDurationEnv("SERVER_SHUTDOWN_TIMEOUT", "30s"),
			Environment:     getEnv("ENVIRONMENT", "development"),
			Version:         getEnv("VERSION", "1.0.0"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getIntEnv("DB_PORT", 5432),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", ""),
			Name:            getEnv("DB_NAME", "taxifleet"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getIntEnv("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getIntEnv("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getDurationEnv("DB_CONN_MAX_LIFETIME", "5m"),
			ConnMaxIdleTime: getDurationEnv("DB_CONN_MAX_IDLE_TIME", "5m"),
			MigrationPath:   getEnv("DB_MIGRATION_PATH", "file://migrations"),
		},
		JWT: JWTConfig{
			Secret:            getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			Expiration:        getDurationEnv("JWT_EXPIRATION", "15m"),
			RefreshExpiration: getDurationEnv("JWT_REFRESH_EXPIRATION", "7d"),
		},
		Permissions: PermissionsConfig{
			Admin:    getIntEnv("JWT_ADMIN_PERMISSION_MASK", 0xFFFFFFFF),
			Owner:    getIntEnv("JWT_OWNER_PERMISSION_MASK", 0xFFFFF), // All except tenant management
			Manager:  getIntEnv("JWT_MANAGER_PERMISSION_MASK", 0x3),   // View and Add reports only
			Mechanic: getIntEnv("JWT_MECHANIC_PERMISSION_MASK", 0x70), // Taxi management
			Driver:   getIntEnv("JWT_DRIVER_PERMISSION_MASK", 0x3),    // View and Add reports
		},
		Security: SecurityConfig{
			BCryptCost:         getIntEnv("BCRYPT_COST", 12),
			RateLimitRPS:       getIntEnv("RATE_LIMIT_RPS", 10),
			RateLimitBurst:     getIntEnv("RATE_LIMIT_BURST", 20),
			CORSAllowedOrigins: getSliceEnv("CORS_ALLOWED_ORIGINS", "*"),
			CORSAllowedMethods: getSliceEnv("CORS_ALLOWED_METHODS", "GET,POST,PUT,DELETE,OPTIONS,PATCH"),
			CORSAllowedHeaders: getSliceEnv("CORS_ALLOWED_HEADERS", "Origin,Content-Type,Accept,Authorization"),
		},
		Logging: LoggingConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			Format:     getEnv("LOG_FORMAT", "json"),
			Output:     getEnv("LOG_OUTPUT", "stdout"),
			MaxSize:    getIntEnv("LOG_MAX_SIZE", 100),
			MaxBackups: getIntEnv("LOG_MAX_BACKUPS", 3),
			MaxAge:     getIntEnv("LOG_MAX_AGE", 28),
			Compress:   getBoolEnv("LOG_COMPRESS", true),
		},
	}

	return config, config.Validate()
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}
	if c.JWT.Secret == "" || c.JWT.Secret == "your-secret-key-change-in-production" {
		if c.Server.Environment == "production" {
			return fmt.Errorf("JWT secret must be set in production")
		}
	}
	return nil
}

// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

// GetAddress returns the server address
func (c *ServerConfig) GetAddress() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// IsDevelopment returns true if the environment is development
func (c *ServerConfig) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if the environment is production
func (c *ServerConfig) IsProduction() bool {
	return c.Environment == "production"
}

// Helper functions for environment variable parsing

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue string) time.Duration {
	value := getEnv(key, defaultValue)
	if duration, err := time.ParseDuration(value); err == nil {
		return duration
	}
	// Fallback to default if parsing fails
	if duration, err := time.ParseDuration(defaultValue); err == nil {
		return duration
	}
	return time.Second * 30 // Ultimate fallback
}

func getSliceEnv(key string, defaultValue string) []string {
	value := getEnv(key, defaultValue)
	return strings.Split(value, ",")
}
