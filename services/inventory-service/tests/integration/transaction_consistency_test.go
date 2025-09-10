package integration_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/services/inventory-service/internal/application/dto"
	appsvc "ecommerce-platform/services/inventory-service/internal/application/services"
	"ecommerce-platform/services/inventory-service/internal/domain/entities"
	"ecommerce-platform/services/inventory-service/internal/domain/ports/repository"
	productRepoImpl "ecommerce-platform/services/inventory-service/internal/infra/repository"
	"ecommerce-platform/services/inventory-service/tests/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestInventoryRepository_TransactionRollback tests that database transactions
// properly rollback when errors occur during stock operations
func TestInventoryRepository_TransactionRollback(t *testing.T) {
	// Arrange
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := productRepoImpl.NewInventoryRepository(db)
	ctx := context.Background()

	// Act - Simulate error in the middle of transaction
	err := repo.WithTransaction(ctx, func(txRepo repository.InventoryRepositoryFacade) error {
		// Create product
		currency, err := valueobjects.NewCurrency("USD")
		require.NoError(t, err)
		price := valueobjects.NewMoney(1000, currency)

		product := entities.NewProduct("test-product-id", "Test Product", price, "https://example.com/image.jpg")
		require.NoError(t, product.Validate())

		if err := txRepo.CreateProduct(ctx, product); err != nil {
			return err
		}

		// Create stock
		stock := entities.NewStock(product.ID, 50, 0)
		if err := txRepo.UpsertStock(ctx, stock); err != nil {
			return err
		}

		// Simulate error
		return errors.New("simulated transaction error")
	})

	// Assert - Transaction should fail and rollback
	require.Error(t, err)
	assert.Contains(t, err.Error(), "simulated transaction error")

	// Verify that product was not created (rollback worked)
	exists, err := repo.ProductExistsByID(ctx, "test-product-id")
	require.NoError(t, err)
	assert.False(t, exists)

	// Verify that stock was not created
	_, err = repo.GetStockByProductID(ctx, "test-product-id")
	require.Error(t, err)
}

// TestInventoryRepository_TransactionSuccess tests that database transactions
// properly commit when all operations succeed
func TestInventoryRepository_TransactionSuccess(t *testing.T) {
	// Arrange
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := productRepoImpl.NewInventoryRepository(db)
	ctx := context.Background()

	// Act - Successful transaction
	err := repo.WithTransaction(ctx, func(txRepo repository.InventoryRepositoryFacade) error {
		// Create product
		currency, err := valueobjects.NewCurrency("USD")
		require.NoError(t, err)
		price := valueobjects.NewMoney(1000, currency)

		product := entities.NewProduct("test-product-id", "Test Product", price, "https://example.com/image.jpg")
		require.NoError(t, product.Validate())

		if err := txRepo.CreateProduct(ctx, product); err != nil {
			return err
		}

		// Create stock
		stock := entities.NewStock(product.ID, 50, 0)
		if err := txRepo.UpsertStock(ctx, stock); err != nil {
			return err
		}

		return nil
	})

	// Assert - Transaction should succeed
	require.NoError(t, err)

	// Verify that product was created
	exists, err := repo.ProductExistsByID(ctx, "test-product-id")
	require.NoError(t, err)
	assert.True(t, exists)

	// Verify that stock was created
	stock, err := repo.GetStockByProductID(ctx, "test-product-id")
	require.NoError(t, err)
	assert.Equal(t, int32(50), stock.AvailableQuantity)
}

// TestInventoryService_StockReservationWithOutboxEvent tests that stock reservation
// and outbox event creation happen atomically in a single transaction
func TestInventoryService_StockReservationWithOutboxEvent(t *testing.T) {
	// Arrange
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := productRepoImpl.NewInventoryRepository(db)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPublisher := mocks.NewMockEventPublisher(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)
	service := appsvc.NewInventoryApplicationService(repo, mockPublisher, mockLogger)

	ctx := context.Background()

	// Create test product
	product := createTestProduct(t, service, "Test Product", 1000, 50)
	require.NotNil(t, product)

	// Act - Reserve stock (this should create both stock reservation and outbox event)
	orderID := "order-123"
	items := []valueobjects.Item{{ProductID: product.ID, Quantity: 30}}

	reserveReq := &dto.ReserveStockRequest{
		OrderID: orderID,
		Items:   convertToDTOItems(items),
	}

	// Setup mock expectations
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	reserveResp, err := service.ReserveStock(ctx, reserveReq)
	require.NoError(t, err)
	assert.True(t, reserveResp.Success)

	// Assert - Both stock reservation and outbox event should exist
	stock, err := service.GetStockByProductID(ctx, product.ID)
	require.NoError(t, err)
	assert.Equal(t, int32(20), stock.AvailableQuantity) // 50 - 30
	assert.Equal(t, int32(30), stock.ReservedQuantity)

	// Verify outbox event was created
	events, err := repo.GetUnprocessedEvents(ctx, 10)
	require.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, "StockReserved", events[0].Type)
	assert.Equal(t, orderID, events[0].AggregateID)
}

// TestInventoryService_BatchStockOperations tests batch stock operations
// to ensure they work correctly with transactions
func TestInventoryService_BatchStockOperations(t *testing.T) {
	// Arrange
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := productRepoImpl.NewInventoryRepository(db)
	ctx := context.Background()

	// Create multiple products
	products := make([]*entities.Product, 3)
	stocks := make([]*entities.Stock, 3)

	for i := 0; i < 3; i++ {
		currency, err := valueobjects.NewCurrency("USD")
		require.NoError(t, err)
		price := valueobjects.NewMoney(1000, currency)

		product := entities.NewProduct(
			fmt.Sprintf("product-%d", i),
			fmt.Sprintf("Test Product %d", i),
			price,
			"https://example.com/image.jpg",
		)
		require.NoError(t, product.Validate())
		products[i] = product

		stock := entities.NewStock(product.ID, 50, 0)
		stocks[i] = stock
	}

	// Act - Create all products and stocks in a single transaction
	err := repo.WithTransaction(ctx, func(txRepo repository.InventoryRepositoryFacade) error {
		// Create all products
		for _, product := range products {
			if err := txRepo.CreateProduct(ctx, product); err != nil {
				return err
			}
		}

		// Create all stocks
		for _, stock := range stocks {
			if err := txRepo.UpsertStock(ctx, stock); err != nil {
				return err
			}
		}

		return nil
	})

	// Assert - All operations should succeed
	require.NoError(t, err)

	// Verify all products were created
	for i := 0; i < 3; i++ {
		exists, err := repo.ProductExistsByID(ctx, fmt.Sprintf("product-%d", i))
		require.NoError(t, err)
		assert.True(t, exists)

		stock, err := repo.GetStockByProductID(ctx, fmt.Sprintf("product-%d", i))
		require.NoError(t, err)
		assert.Equal(t, int32(50), stock.AvailableQuantity)
	}
}
