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

	"ecommerce-platform/services/payment-service/internal/infra/migration"
)

// SetupTestDatabase creates a test database with PostgreSQL container
func SetupTestDatabase(t *testing.T, ctx context.Context) (*gorm.DB, func()) {
	// Start PostgreSQL container
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15.4",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "payment_test",
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	// Get connection string
	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	dsn := "host=" + host + " port=" + port.Port() + " user=test password=test dbname=payment_test sslmode=disable"

	// Connect to database with retry
	var db *gorm.DB
	for i := 0; i < 15; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			// Test the connection
			sqlDB, err := db.DB()
			if err == nil {
				if pingErr := sqlDB.Ping(); pingErr == nil {
					break
				}
			}
		}
		if i < 14 {
			time.Sleep(1 * time.Second)
		}
	}
	require.NoError(t, err, "Failed to connect to database after retries")

	// Run migrations
	migrationService := migration.NewMigrationService(db)
	err = migrationService.Migrate(ctx)
	require.NoError(t, err)

	// Cleanup function
	cleanup := func() {
		if container != nil {
			container.Terminate(ctx)
		}
	}

	return db, cleanup
}

// CleanDatabase clears all tables in test database
func CleanDatabase(t *testing.T, db *gorm.DB) {
	// Clear tables in correct order (considering foreign keys)
	tables := []string{"outbox_events"}

	for _, table := range tables {
		err := db.Exec("DELETE FROM " + table).Error
		require.NoError(t, err)
	}
}
