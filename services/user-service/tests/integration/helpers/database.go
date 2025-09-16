package helpers

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"

	"ecommerce-platform/pkg/redisclient"
	"ecommerce-platform/services/user-service/internal/infra/migration"
)

// SetupTestDatabase creates test databases with PostgreSQL and Redis containers
func SetupTestDatabase(t *testing.T, ctx context.Context) (*gorm.DB, *redisclient.Client, func()) {
	// Start PostgreSQL container
	postgresReq := testcontainers.ContainerRequest{
		Image:        "postgres:15.4",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "user_service_test",
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections"),
	}

	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: postgresReq,
		Started:          true,
	})
	require.NoError(t, err)

	// Start Redis container
	redisReq := testcontainers.ContainerRequest{
		Image:        "redis:7.0-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: redisReq,
		Started:          true,
	})
	require.NoError(t, err)

	// Get PostgreSQL connection string
	postgresHost, err := postgresContainer.Host(ctx)
	require.NoError(t, err)

	postgresPort, err := postgresContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	postgresDSN := "host=" + postgresHost + " port=" + postgresPort.Port() + " user=test password=test dbname=user_service_test sslmode=disable"

	// Connect to PostgreSQL with retry
	var db *gorm.DB
	for i := 0; i < 10; i++ {
		db, err = gorm.Open(postgres.Open(postgresDSN), &gorm.Config{})
		if err == nil {
			break
		}
		if i < 9 {
			time.Sleep(500 * time.Millisecond)
		}
	}
	require.NoError(t, err, "Failed to connect to PostgreSQL after retries")

	// Run migrations
	migrationService := migration.NewMigrationService(db)
	err = migrationService.Migrate(ctx)
	require.NoError(t, err)

	// Disable GORM logging in tests to reduce noise
	db.Logger = gormLogger.Default.LogMode(gormLogger.Silent)

	// Get Redis connection string
	redisHost, err := redisContainer.Host(ctx)
	require.NoError(t, err)

	redisPort, err := redisContainer.MappedPort(ctx, "6379")
	require.NoError(t, err)

	redisAddr := redisHost + ":" + redisPort.Port()

	// Connect to Redis
	redisClient := redisclient.New(redisAddr, "", 0)

	// Test Redis connection
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = redisClient.Ping(ctx)
	require.NoError(t, err, "Failed to connect to Redis")

	// Cleanup function
	cleanup := func() {
		if postgresContainer != nil {
			postgresContainer.Terminate(ctx)
		}
		if redisContainer != nil {
			redisContainer.Terminate(ctx)
		}
	}

	return db, redisClient, cleanup
}

// CleanDatabase clears all tables in test database
func CleanDatabase(t *testing.T, db *gorm.DB) {
	// Clear tables in correct order (considering foreign keys)
	tables := []string{"users"}

	for _, table := range tables {
		err := db.Exec("TRUNCATE TABLE " + table + " RESTART IDENTITY CASCADE").Error
		require.NoError(t, err)
	}
}

// CleanRedis clears all keys in test Redis
func CleanRedis(t *testing.T, ctx context.Context, redisClient *redisclient.Client) {
	// Flush entire Redis database for clean test state
	err := redisClient.FlushDB(ctx)
	require.NoError(t, err)
}
