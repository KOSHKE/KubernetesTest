package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// BaseConfig contains common configuration for all services
type BaseConfig struct {
	// Service identification
	ServiceName string
	Environment string

	// Server configuration
	Port        string
	MetricsPort string

	// Database configuration
	Database DatabaseConfig

	// Kafka configuration
	Kafka KafkaConfig

	// Monitoring configuration
	Monitoring MonitoringConfig

	// Logging configuration
	Logging LoggingConfig
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Host        string
	Port        string
	Name        string
	User        string
	Password    string
	SSLMode     string
	MaxConns    int
	Timeout     time.Duration
	AutoMigrate bool
}

// KafkaConfig holds Kafka-related configuration
type KafkaConfig struct {
	Brokers           []string
	ClientID          string
	GroupID           string
	AutoOffsetReset   string
	SessionTimeout    time.Duration
	HeartbeatInterval time.Duration
	MaxPollRecords    int
	MaxPollInterval   time.Duration
}

// MonitoringConfig holds monitoring-related configuration
type MonitoringConfig struct {
	Enabled       bool
	MetricsPath   string
	HealthPath    string
	ReadinessPath string
}

// LoggingConfig holds logging-related configuration
type LoggingConfig struct {
	Level      string
	Format     string
	OutputPath string
	MaxSize    int
	MaxAge     int
	MaxBackups int
	Compress   bool
}

// LoadBaseConfig loads common configuration from environment variables
func LoadBaseConfig(serviceName string) *BaseConfig {
	cfg := &BaseConfig{}

	// Set service defaults
	cfg.ServiceName = serviceName
	cfg.Environment = getEnv("ENVIRONMENT", "development")

	// Load server configuration
	cfg.loadServerConfig()

	// Load database configuration
	cfg.loadDatabaseConfig()

	// Load Kafka configuration
	cfg.loadKafkaConfig()

	// Load monitoring configuration
	cfg.loadMonitoringConfig()

	// Load logging configuration
	cfg.loadLoggingConfig()

	return cfg
}

// loadServerConfig loads server-related configuration
func (cfg *BaseConfig) loadServerConfig() {
	cfg.Port = getEnv("PORT", "8080")
	cfg.MetricsPort = getEnv("METRICS_PORT", "9090")
}

// loadDatabaseConfig loads database-related configuration
func (cfg *BaseConfig) loadDatabaseConfig() {
	cfg.Database.Host = getEnv("DB_HOST", "localhost")
	cfg.Database.Port = getEnv("DB_PORT", "5432")
	cfg.Database.Name = getEnv("DB_NAME", "default")
	cfg.Database.User = getEnv("DB_USER", "user")
	cfg.Database.Password = getEnv("DB_PASSWORD", "password")
	cfg.Database.SSLMode = getEnv("DB_SSLMODE", "disable")
	cfg.Database.MaxConns = getEnvAsInt("DB_MAX_CONNS", 10)
	cfg.Database.Timeout = getEnvAsDuration("DB_TIMEOUT", 5*time.Second)
	cfg.Database.AutoMigrate = getEnvAsBool("DB_AUTO_MIGRATE", false)
}

// loadKafkaConfig loads Kafka-related configuration
func (cfg *BaseConfig) loadKafkaConfig() {
	// Support both KAFKA_BROKERS and KAFKA_BOOTSTRAP_SERVERS
	kafkaBrokers := getEnv("KAFKA_BROKERS", "")
	kafkaBootstrapServers := getEnv("KAFKA_BOOTSTRAP_SERVERS", "")

	if kafkaBrokers != "" {
		cfg.Kafka.Brokers = strings.Split(kafkaBrokers, ",")
	} else if kafkaBootstrapServers != "" {
		cfg.Kafka.Brokers = strings.Split(kafkaBootstrapServers, ",")
	} else {
		cfg.Kafka.Brokers = []string{"kafka:9092"}
	}

	cfg.Kafka.ClientID = getEnv("KAFKA_CLIENT_ID", cfg.ServiceName)
	cfg.Kafka.GroupID = getEnv("KAFKA_GROUP_ID", cfg.ServiceName)
	cfg.Kafka.AutoOffsetReset = getEnv("KAFKA_AUTO_OFFSET_RESET", "earliest")
	cfg.Kafka.SessionTimeout = getEnvAsDuration("KAFKA_SESSION_TIMEOUT", 30*time.Second)
	cfg.Kafka.HeartbeatInterval = getEnvAsDuration("KAFKA_HEARTBEAT_INTERVAL", 3*time.Second)
	cfg.Kafka.MaxPollRecords = getEnvAsInt("KAFKA_MAX_POLL_RECORDS", 500)
	cfg.Kafka.MaxPollInterval = getEnvAsDuration("KAFKA_MAX_POLL_INTERVAL", 300*time.Second)
}

// loadMonitoringConfig loads monitoring-related configuration
func (cfg *BaseConfig) loadMonitoringConfig() {
	cfg.Monitoring.Enabled = getEnvAsBool("MONITORING_ENABLED", true)
	cfg.Monitoring.MetricsPath = getEnv("MONITORING_METRICS_PATH", "/metrics")
	cfg.Monitoring.HealthPath = getEnv("MONITORING_HEALTH_PATH", "/health")
	cfg.Monitoring.ReadinessPath = getEnv("MONITORING_READINESS_PATH", "/ready")
}

// loadLoggingConfig loads logging-related configuration
func (cfg *BaseConfig) loadLoggingConfig() {
	cfg.Logging.Level = getEnv("LOG_LEVEL", "info")
	cfg.Logging.Format = getEnv("LOG_FORMAT", "json")
	cfg.Logging.OutputPath = getEnv("LOG_OUTPUT_PATH", "stdout")
	cfg.Logging.MaxSize = getEnvAsInt("LOG_MAX_SIZE", 100)
	cfg.Logging.MaxAge = getEnvAsInt("LOG_MAX_AGE", 30)
	cfg.Logging.MaxBackups = getEnvAsInt("LOG_MAX_BACKUPS", 10)
	cfg.Logging.Compress = getEnvAsBool("LOG_COMPRESS", true)
}

// Validate validates the base configuration
func (cfg *BaseConfig) Validate() error {
	if cfg.Port == "" {
		return fmt.Errorf("PORT is required")
	}

	if cfg.MetricsPort == "" {
		return fmt.Errorf("METRICS_PORT is required")
	}

	if len(cfg.Kafka.Brokers) == 0 {
		return fmt.Errorf("at least one Kafka broker is required")
	}

	return nil
}

// GetDatabaseDSN returns the database connection string
func (cfg *BaseConfig) GetDatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
		int(cfg.Database.Timeout.Seconds()),
	)
}

// GetKafkaBrokers returns Kafka brokers as comma-separated string
func (cfg *BaseConfig) GetKafkaBrokers() string {
	return strings.Join(cfg.Kafka.Brokers, ",")
}

// IsDevelopment returns true if running in development environment
func (cfg *BaseConfig) IsDevelopment() bool {
	return cfg.Environment == "development"
}

// IsProduction returns true if running in production environment
func (cfg *BaseConfig) IsProduction() bool {
	return cfg.Environment == "production"
}

// IsTest returns true if running in test environment
func (cfg *BaseConfig) IsTest() bool {
	return cfg.Environment == "test"
}

// Helper functions for environment variable parsing

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
