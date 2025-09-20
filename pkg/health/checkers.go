package health

import (
	"context"
	"fmt"

	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"gorm.io/gorm"
)

// DatabaseChecker checks database connectivity
type DatabaseChecker struct {
	db *gorm.DB
}

// NewDatabaseChecker creates a new database health checker
func NewDatabaseChecker(db *gorm.DB) *DatabaseChecker {
	return &DatabaseChecker{db: db}
}

func (c *DatabaseChecker) Name() string {
	return "database"
}

func (c *DatabaseChecker) Check(ctx context.Context) error {
	if c.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	sqlDB, err := c.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	return sqlDB.PingContext(ctx)
}

// ConsumerChecker checks if consumers are running
type ConsumerChecker struct {
	name      string
	isRunning func() bool
}

// NewConsumerChecker creates a new consumer health checker
func NewConsumerChecker(name string, isRunning func() bool) *ConsumerChecker {
	return &ConsumerChecker{
		name:      name,
		isRunning: isRunning,
	}
}

func (c *ConsumerChecker) Name() string {
	return fmt.Sprintf("consumer_%s", c.name)
}

func (c *ConsumerChecker) Check(ctx context.Context) error {
	if !c.isRunning() {
		return fmt.Errorf("consumer %s is not running", c.name)
	}
	return nil
}

// PublisherChecker checks if publisher is available
type PublisherChecker struct {
	name      string
	isHealthy func(ctx context.Context) error
}

// NewPublisherChecker creates a new publisher health checker
func NewPublisherChecker(name string, isHealthy func(ctx context.Context) error) *PublisherChecker {
	return &PublisherChecker{
		name:      name,
		isHealthy: isHealthy,
	}
}

func (c *PublisherChecker) Name() string {
	return fmt.Sprintf("publisher_%s", c.name)
}

func (c *PublisherChecker) Check(ctx context.Context) error {
	if c.isHealthy == nil {
		return nil // No health check function provided
	}
	return c.isHealthy(ctx)
}

// OutboxChecker checks if outbox publisher is running
type OutboxChecker struct {
	isRunning func() bool
}

// NewOutboxChecker creates a new outbox health checker
func NewOutboxChecker(isRunning func() bool) *OutboxChecker {
	return &OutboxChecker{
		isRunning: isRunning,
	}
}

func (c *OutboxChecker) Name() string {
	return "outbox_publisher"
}

func (c *OutboxChecker) Check(ctx context.Context) error {
	if !c.isRunning() {
		return fmt.Errorf("outbox publisher is not running")
	}
	return nil
}

// RedisChecker checks Redis connectivity using our redisclient wrapper
type RedisChecker struct {
	client interface {
		Ping(ctx context.Context) error
	}
}

// NewRedisChecker creates a new Redis health checker
// Accepts any client that implements Ping(ctx context.Context) error
func NewRedisChecker(client interface {
	Ping(ctx context.Context) error
}) *RedisChecker {
	return &RedisChecker{client: client}
}

func (c *RedisChecker) Name() string {
	return "redis"
}

func (c *RedisChecker) Check(ctx context.Context) error {
	if c.client == nil {
		return fmt.Errorf("redis client is nil")
	}

	// Use our client's Ping method
	if err := c.client.Ping(ctx); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	return nil
}

// GRPCHealthChecker manages gRPC health status and acts as a health checker
type GRPCHealthChecker struct {
	grpcHealth    *health.Server
	healthManager interface {
		IsHealthy(ctx context.Context) bool
	}
}

// NewGRPCHealthChecker creates a new gRPC health checker that syncs with health manager
func NewGRPCHealthChecker(healthManager interface {
	IsHealthy(ctx context.Context) bool
}) *GRPCHealthChecker {
	grpcHealth := health.NewServer()
	// Initially set as not serving
	grpcHealth.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)

	return &GRPCHealthChecker{
		grpcHealth:    grpcHealth,
		healthManager: healthManager,
	}
}

// GetGRPCHealthServer returns the gRPC health server for registration
func (c *GRPCHealthChecker) GetGRPCHealthServer() *health.Server {
	return c.grpcHealth
}

func (c *GRPCHealthChecker) Name() string {
	return "grpc_health_sync"
}

func (c *GRPCHealthChecker) Check(ctx context.Context) error {
	// Update gRPC health status based on overall health
	if c.healthManager.IsHealthy(ctx) {
		c.grpcHealth.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	} else {
		c.grpcHealth.SetServingStatus("", healthpb.HealthCheckResponse_NOT_SERVING)
	}
	return nil // This checker never fails, it just syncs status
}
