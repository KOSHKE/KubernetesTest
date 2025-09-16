package helpers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"ecommerce-platform/services/inventory-service/internal/infra/migration"
)

// SetupTestDatabase creates a test database with PostgreSQL container
func SetupTestDatabase(t *testing.T, ctx context.Context) (*gorm.DB, func()) {
	// Start PostgreSQL container
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15.4",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "inventory_test",
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

	dsn := "host=" + host + " port=" + port.Port() + " user=test password=test dbname=inventory_test sslmode=disable"

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

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
	tables := []string{"outbox_events", "stocks", "products"}

	for _, table := range tables {
		err := db.Exec("DELETE FROM " + table).Error
		require.NoError(t, err)
	}
}
