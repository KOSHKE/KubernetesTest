package integration_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/services/inventory-service/internal/application/dto"
	appsvc "ecommerce-platform/services/inventory-service/internal/application/services"
	productRepoImpl "ecommerce-platform/services/inventory-service/internal/infra/repository"
	"ecommerce-platform/services/inventory-service/tests/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestInventoryService_StockReservationFlow tests the complete stock reservation flow
// including reserve, release, and commit operations with database persistence
func TestInventoryService_StockReservationFlow(t *testing.T) {
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

	// Create test product with initial stock
	product := createTestProduct(t, service, "Test Product", 1000, 50)
	require.NotNil(t, product)

	// Verify initial stock state
	initialStock, err := service.GetStockByProductID(ctx, product.ID)
	require.NoError(t, err)
	assert.Equal(t, int32(50), initialStock.AvailableQuantity)
	assert.Equal(t, int32(0), initialStock.ReservedQuantity)

	// Act & Assert - Reserve Stock
	orderID := "order-123"
	items := []valueobjects.Item{{ProductID: product.ID, Quantity: 30}}

	reserveReq := &dto.ReserveStockRequest{
		OrderID: orderID,
		Items:   convertToDTOItems(items),
	}

	// Setup mock expectations for reserve
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	reserveResp, err := service.ReserveStock(ctx, reserveReq)
	require.NoError(t, err)
	assert.True(t, reserveResp.Success)
	assert.Equal(t, orderID, reserveResp.OrderID)
	assert.Len(t, reserveResp.ReservedItems, 1)
	assert.Equal(t, product.ID, reserveResp.ReservedItems[0])

	// Verify stock state after reservation
	stockAfterReserve, err := service.GetStockByProductID(ctx, product.ID)
	require.NoError(t, err)
	assert.Equal(t, int32(20), stockAfterReserve.AvailableQuantity) // 50 - 30
	assert.Equal(t, int32(30), stockAfterReserve.ReservedQuantity)

	// Verify outbox event was created
	events, err := repo.GetUnprocessedEvents(ctx, 10)
	require.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, "StockReserved", events[0].Type)
	assert.Equal(t, orderID, events[0].AggregateID)

	// Act & Assert - Release Stock
	releaseReq := &dto.ReleaseStockRequest{
		OrderID: orderID,
		Items:   convertToDTOItems(items),
	}

	releaseResp, err := service.ReleaseStock(ctx, releaseReq)
	require.NoError(t, err)
	assert.True(t, releaseResp.Success)
	assert.Equal(t, orderID, releaseResp.OrderID)

	// Verify stock state after release
	stockAfterRelease, err := service.GetStockByProductID(ctx, product.ID)
	require.NoError(t, err)
	assert.Equal(t, int32(50), stockAfterRelease.AvailableQuantity) // 20 + 30
	assert.Equal(t, int32(0), stockAfterRelease.ReservedQuantity)

	// Verify additional outbox event was created
	events, err = repo.GetUnprocessedEvents(ctx, 10)
	require.NoError(t, err)
	assert.Len(t, events, 2) // Reserve + Release events

	// Act & Assert - Reserve and Commit (without release)
	reserveReq2 := &dto.ReserveStockRequest{
		OrderID: "order-456",
		Items:   convertToDTOItems(items),
	}

	reserveResp2, err := service.ReserveStock(ctx, reserveReq2)
	require.NoError(t, err)
	assert.True(t, reserveResp2.Success)

	// Verify stock is reserved
	stockAfterReserve2, err := service.GetStockByProductID(ctx, product.ID)
	require.NoError(t, err)
	assert.Equal(t, int32(20), stockAfterReserve2.AvailableQuantity)
	assert.Equal(t, int32(30), stockAfterReserve2.ReservedQuantity)

	// Commit the reserved stock
	commitReq := &dto.CommitStockRequest{
		OrderID: "order-456",
		Items:   convertToDTOItems(items),
	}

	commitResp, err := service.CommitStock(ctx, commitReq)
	require.NoError(t, err)
	assert.True(t, commitResp.Success)
	assert.Equal(t, "order-456", commitResp.OrderID)

	// Verify stock state after commit (reserved quantity should be reduced)
	stockAfterCommit, err := service.GetStockByProductID(ctx, product.ID)
	require.NoError(t, err)
	assert.Equal(t, int32(20), stockAfterCommit.AvailableQuantity) // unchanged
	assert.Equal(t, int32(0), stockAfterCommit.ReservedQuantity)   // 30 - 30

	// Verify all outbox events were created
	events, err = repo.GetUnprocessedEvents(ctx, 10)
	require.NoError(t, err)
	assert.Len(t, events, 3) // Reserve + Release + Reserve + Commit events
}

// TestInventoryService_InsufficientStock tests stock reservation with insufficient available stock
func TestInventoryService_InsufficientStock(t *testing.T) {
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

	// Create test product with limited stock
	product := createTestProduct(t, service, "Test Product", 1000, 10)
	require.NotNil(t, product)

	// Act - Try to reserve more stock than available
	orderID := "order-123"
	items := []valueobjects.Item{{ProductID: product.ID, Quantity: 20}} // More than available (10)

	reserveReq := &dto.ReserveStockRequest{
		OrderID: orderID,
		Items:   convertToDTOItems(items),
	}

	// Setup mock expectations
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	// Assert - Should fail with insufficient stock error
	_, err := service.ReserveStock(ctx, reserveReq)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient stock")

	// Verify stock state remains unchanged
	stock, err := service.GetStockByProductID(ctx, product.ID)
	require.NoError(t, err)
	assert.Equal(t, int32(10), stock.AvailableQuantity)
	assert.Equal(t, int32(0), stock.ReservedQuantity)

	// Verify no outbox events were created
	events, err := repo.GetUnprocessedEvents(ctx, 10)
	require.NoError(t, err)
	assert.Len(t, events, 0)
}

// TestInventoryService_ConcurrentStockReservation tests concurrent stock reservations
func TestInventoryService_ConcurrentStockReservation(t *testing.T) {
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

	// Create test product with limited stock
	product := createTestProduct(t, service, "Test Product", 1000, 100)
	require.NotNil(t, product)

	// Act - 10 goroutines try to reserve 20 units each (total 200, but only 100 available)
	const numGoroutines = 10
	const quantityPerReservation = 20

	var wg sync.WaitGroup
	successCount := int32(0)
	errorCount := int32(0)
	errors := make(chan error, numGoroutines)

	// Setup mock expectations
	mockLogger.EXPECT().Error(gomock.Any(), gomock.Any()).AnyTimes()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(orderNum int) {
			defer wg.Done()

			items := []valueobjects.Item{{ProductID: product.ID, Quantity: quantityPerReservation}}
			reserveReq := &dto.ReserveStockRequest{
				OrderID: fmt.Sprintf("order-%d", orderNum),
				Items:   convertToDTOItems(items),
			}

			reserveResp, err := service.ReserveStock(ctx, reserveReq)
			if err != nil {
				atomic.AddInt32(&errorCount, 1)
				errors <- err
			} else {
				atomic.AddInt32(&successCount, 1)
				_ = reserveResp // Use the response to avoid unused variable
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Assert - Only 5 operations should succeed (100/20 = 5)
	assert.Equal(t, int32(5), successCount)
	assert.Equal(t, int32(5), errorCount)

	// Verify final stock state
	stock, err := service.GetStockByProductID(ctx, product.ID)
	require.NoError(t, err)
	assert.Equal(t, int32(0), stock.AvailableQuantity)  // All stock reserved
	assert.Equal(t, int32(100), stock.ReservedQuantity) // 5 * 20 = 100

	// Verify outbox events were created for successful reservations
	events, err := repo.GetUnprocessedEvents(ctx, 10)
	require.NoError(t, err)
	assert.Len(t, events, 5) // 5 successful reservations
}
