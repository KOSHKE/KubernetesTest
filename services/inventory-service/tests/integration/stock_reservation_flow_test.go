package integration

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/mock/gomock"
	gormpg "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/services/inventory-service/internal/application/usecases"
	"ecommerce-platform/services/inventory-service/internal/infra/migration"
	"ecommerce-platform/services/inventory-service/internal/infra/repository"
	"ecommerce-platform/services/inventory-service/tests/mocks"
)

func setupTestDB(t *testing.T, ctx context.Context) (*gorm.DB, func()) {
	postgresC, err := tcpostgres.Run(ctx, "postgres:15.4",
		tcpostgres.WithDatabase("inventory_test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err)

	connStr, err := postgresC.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := gorm.Open(gormpg.Open(connStr), &gorm.Config{})
	require.NoError(t, err)

	migrator := migration.NewMigrationService(db)
	require.NoError(t, migrator.RunMigrations(ctx))
	require.NoError(t, migrator.SeedData(ctx))

	cleanup := func() { _ = postgresC.Terminate(ctx) }

	return db, cleanup
}

func TestStockReservationFlow(t *testing.T) {
	ctx := context.Background()
	db, cleanup := setupTestDB(t, ctx)
	defer cleanup()

	repo := repository.NewInventoryRepository(db)

	// Create mock outbox service
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	outboxSvc := mocks.NewMockService(ctrl)

	reserveUC := usecases.NewReserveStockUseCase(repo, outboxSvc)
	commitUC := usecases.NewCommitStockUseCase(repo, outboxSvc)
	releaseUC := usecases.NewReleaseStockUseCase(repo, outboxSvc)

	t.Run("successful reservation", func(t *testing.T) {
		orderID := "order-123"
		items := []valueobjects.Item{{ProductID: "prod-1", Quantity: 2}}

		// Setup expectations for outbox service
		outboxSvc.EXPECT().SaveEvent(ctx, gomock.Any()).Return(nil).Times(1)

		err := reserveUC.Execute(ctx, orderID, items)
		require.NoError(t, err)

		stock, err := repo.GetStockByProductID(ctx, "prod-1")
		require.NoError(t, err)
		require.Equal(t, int32(1), stock.AvailableQuantity)
		require.Equal(t, int32(2), stock.ReservedQuantity)
	})

	t.Run("concurrent reservation", func(t *testing.T) {
		order1, order2 := "o1", "o2"
		items := []valueobjects.Item{{ProductID: "prod-2", Quantity: 1}}

		// Setup expectations - only one call should be successful
		outboxSvc.EXPECT().SaveEvent(ctx, gomock.Any()).Return(nil).Times(1)

		var wg sync.WaitGroup
		errs := make(chan error, 2)

		wg.Add(2)
		go func() { defer wg.Done(); errs <- reserveUC.Execute(ctx, order1, items) }()
		go func() { defer wg.Done(); errs <- reserveUC.Execute(ctx, order2, items) }()
		wg.Wait()
		close(errs)

		success := 0
		for err := range errs {
			if err == nil {
				success++
			}
		}
		require.Equal(t, 1, success)
	})

	t.Run("commit reserved stock", func(t *testing.T) {
		orderID := "order-commit"
		items := []valueobjects.Item{{ProductID: "prod-3", Quantity: 1}}

		// Setup expectations for two SaveEvent calls
		outboxSvc.EXPECT().SaveEvent(ctx, gomock.Any()).Return(nil).Times(2)

		// First reserve the stock
		err := reserveUC.Execute(ctx, orderID, items)
		require.NoError(t, err)

		// Check that stock is reserved
		stock, err := repo.GetStockByProductID(ctx, "prod-3")
		require.NoError(t, err)
		require.Equal(t, int32(1), stock.AvailableQuantity) // 2 - 1 = 1
		require.Equal(t, int32(1), stock.ReservedQuantity)  // 0 + 1 = 1

		// Commit the reserved stock
		err = commitUC.Execute(ctx, orderID, items)
		require.NoError(t, err)

		// Check that stock is committed
		stock, err = repo.GetStockByProductID(ctx, "prod-3")
		require.NoError(t, err)
		require.Equal(t, int32(1), stock.AvailableQuantity) // Remains 1
		require.Equal(t, int32(0), stock.ReservedQuantity)  // 1 - 1 = 0 (committed)
	})

	t.Run("release reserved stock", func(t *testing.T) {
		orderID := "order-release"
		items := []valueobjects.Item{{ProductID: "prod-3", Quantity: 1}}

		// Setup expectations for two SaveEvent calls
		outboxSvc.EXPECT().SaveEvent(ctx, gomock.Any()).Return(nil).Times(2)

		// First reserve the stock
		err := reserveUC.Execute(ctx, orderID, items)
		require.NoError(t, err)

		// Check that stock is reserved
		stock, err := repo.GetStockByProductID(ctx, "prod-3")
		require.NoError(t, err)
		require.Equal(t, int32(0), stock.AvailableQuantity) // 1 - 1 = 0
		require.Equal(t, int32(1), stock.ReservedQuantity)  // 0 + 1 = 1

		// Release the reserved stock
		err = releaseUC.Execute(ctx, orderID, items)
		require.NoError(t, err)

		// Check that stock is released
		stock, err = repo.GetStockByProductID(ctx, "prod-3")
		require.NoError(t, err)
		require.Equal(t, int32(1), stock.AvailableQuantity) // 0 + 1 = 1 (returned to available)
		require.Equal(t, int32(0), stock.ReservedQuantity)  // 1 - 1 = 0 (released)
	})

	t.Run("insufficient stock for reservation", func(t *testing.T) {
		orderID := "order-insufficient"
		items := []valueobjects.Item{{ProductID: "prod-1", Quantity: 10}} // More than available

		// Don't expect SaveEvent calls on error
		outboxSvc.EXPECT().SaveEvent(ctx, gomock.Any()).Return(nil).Times(0)

		err := reserveUC.Execute(ctx, orderID, items)
		require.Error(t, err)
		require.Contains(t, err.Error(), "insufficient stock")

		// Check that stock remains unchanged
		stock, err := repo.GetStockByProductID(ctx, "prod-1")
		require.NoError(t, err)
		require.Equal(t, int32(1), stock.AvailableQuantity) // Remains unchanged
		require.Equal(t, int32(2), stock.ReservedQuantity)  // Remains unchanged
	})

	t.Run("product not found", func(t *testing.T) {
		orderID := "order-not-found"
		items := []valueobjects.Item{{ProductID: "non-existent", Quantity: 1}}

		// Don't expect SaveEvent calls on error
		outboxSvc.EXPECT().SaveEvent(ctx, gomock.Any()).Return(nil).Times(0)

		err := reserveUC.Execute(ctx, orderID, items)
		require.Error(t, err)
		require.Contains(t, err.Error(), "product not found")
	})
}
