package usecases_test

import (
	"context"
	"testing"

	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/services/inventory-service/internal/application/usecases"
	"ecommerce-platform/services/inventory-service/internal/domain/entities"
	"ecommerce-platform/services/inventory-service/internal/domain/ports/repository"
	"ecommerce-platform/services/inventory-service/tests/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestReserveStockUseCase_Execute_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockInventoryRepositoryFacade(ctrl)
	mockOutbox := mocks.NewMockService(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	useCase := usecases.NewReserveStockUseCase(mockRepo, mockOutbox, mockLogger)

	ctx := context.Background()
	orderID := "order-123"
	items := []valueobjects.Item{
		{ProductID: "product-1", Quantity: 10},
		{ProductID: "product-2", Quantity: 5},
	}

	// Setup mocks
	stocks := map[string]*entities.Stock{
		"product-1": entities.NewStock("product-1", 50, 0), // 50 available
		"product-2": entities.NewStock("product-2", 30, 0), // 30 available
	}

	mockRepo.EXPECT().WithTransaction(ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(repository.InventoryRepositoryFacade) error) error {
			return fn(mockRepo)
		},
	)
	mockRepo.EXPECT().GetStocksByProductIDs(ctx, []string{"product-1", "product-2"}, true).Return(stocks, nil)
	mockRepo.EXPECT().UpsertStocks(ctx, gomock.Any()).Return(nil)
	mockOutbox.EXPECT().SaveEvent(ctx, gomock.Any()).Return(nil)

	// Act
	err := useCase.Execute(ctx, orderID, items)

	// Assert
	assert.NoError(t, err)
}

func TestReserveStockUseCase_Execute_ProductNotFound(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockInventoryRepositoryFacade(ctrl)
	mockOutbox := mocks.NewMockService(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	useCase := usecases.NewReserveStockUseCase(mockRepo, mockOutbox, mockLogger)

	ctx := context.Background()
	orderID := "order-123"
	items := []valueobjects.Item{
		{ProductID: "product-1", Quantity: 10},
	}

	// Setup mocks - product not found
	stocks := map[string]*entities.Stock{}

	mockRepo.EXPECT().WithTransaction(ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(repository.InventoryRepositoryFacade) error) error {
			return fn(mockRepo)
		},
	)
	mockRepo.EXPECT().GetStocksByProductIDs(ctx, []string{"product-1"}, true).Return(stocks, nil)
	mockLogger.EXPECT().Warn("product not found", "productID", "product-1")

	// Act
	err := useCase.Execute(ctx, orderID, items)

	// Assert
	assert.Error(t, err)
}

func TestReserveStockUseCase_Execute_InsufficientStock(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockInventoryRepositoryFacade(ctrl)
	mockOutbox := mocks.NewMockService(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	useCase := usecases.NewReserveStockUseCase(mockRepo, mockOutbox, mockLogger)

	ctx := context.Background()
	orderID := "order-123"
	items := []valueobjects.Item{
		{ProductID: "product-1", Quantity: 100}, // Requesting 100 but only 50 available
	}

	// Setup mocks
	stocks := map[string]*entities.Stock{
		"product-1": entities.NewStock("product-1", 50, 0), // Only 50 available
	}

	mockRepo.EXPECT().WithTransaction(ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(repository.InventoryRepositoryFacade) error) error {
			return fn(mockRepo)
		},
	)
	mockRepo.EXPECT().GetStocksByProductIDs(ctx, []string{"product-1"}, true).Return(stocks, nil)
	mockLogger.EXPECT().Warn("insufficient stock", "productID", "product-1", "requested", int32(100), "available", int32(50))

	// Act
	err := useCase.Execute(ctx, orderID, items)

	// Assert
	assert.Error(t, err)
}

func TestReserveStockUseCase_Execute_GetStocksError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockInventoryRepositoryFacade(ctrl)
	mockOutbox := mocks.NewMockService(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	useCase := usecases.NewReserveStockUseCase(mockRepo, mockOutbox, mockLogger)

	ctx := context.Background()
	orderID := "order-123"
	items := []valueobjects.Item{
		{ProductID: "product-1", Quantity: 10},
	}

	// Setup mocks - repository error
	mockRepo.EXPECT().WithTransaction(ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(repository.InventoryRepositoryFacade) error) error {
			return fn(mockRepo)
		},
	)
	mockRepo.EXPECT().GetStocksByProductIDs(ctx, []string{"product-1"}, true).Return(nil, assert.AnError)
	mockLogger.EXPECT().Error("failed to get stocks", "productIDs", []string{"product-1"}, "error", assert.AnError)

	// Act
	err := useCase.Execute(ctx, orderID, items)

	// Assert
	assert.Error(t, err)
}

func TestReserveStockUseCase_Execute_UpsertStocksError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockInventoryRepositoryFacade(ctrl)
	mockOutbox := mocks.NewMockService(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	useCase := usecases.NewReserveStockUseCase(mockRepo, mockOutbox, mockLogger)

	ctx := context.Background()
	orderID := "order-123"
	items := []valueobjects.Item{
		{ProductID: "product-1", Quantity: 10},
	}

	// Setup mocks
	stocks := map[string]*entities.Stock{
		"product-1": entities.NewStock("product-1", 50, 0),
	}

	mockRepo.EXPECT().WithTransaction(ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(repository.InventoryRepositoryFacade) error) error {
			return fn(mockRepo)
		},
	)
	mockRepo.EXPECT().GetStocksByProductIDs(ctx, []string{"product-1"}, true).Return(stocks, nil)
	mockRepo.EXPECT().UpsertStocks(ctx, gomock.Any()).Return(assert.AnError)
	mockLogger.EXPECT().Error("failed to save stock reservations", "error", assert.AnError)

	// Act
	err := useCase.Execute(ctx, orderID, items)

	// Assert
	assert.Error(t, err)
}

func TestReserveStockUseCase_Execute_OutboxError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockInventoryRepositoryFacade(ctrl)
	mockOutbox := mocks.NewMockService(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	useCase := usecases.NewReserveStockUseCase(mockRepo, mockOutbox, mockLogger)

	ctx := context.Background()
	orderID := "order-123"
	items := []valueobjects.Item{
		{ProductID: "product-1", Quantity: 10},
	}

	// Setup mocks
	stocks := map[string]*entities.Stock{
		"product-1": entities.NewStock("product-1", 50, 0),
	}

	mockRepo.EXPECT().WithTransaction(ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(repository.InventoryRepositoryFacade) error) error {
			return fn(mockRepo)
		},
	)
	mockRepo.EXPECT().GetStocksByProductIDs(ctx, []string{"product-1"}, true).Return(stocks, nil)
	mockRepo.EXPECT().UpsertStocks(ctx, gomock.Any()).Return(nil)
	mockOutbox.EXPECT().SaveEvent(ctx, gomock.Any()).Return(assert.AnError)
	mockLogger.EXPECT().Error("failed to save event to outbox", "orderID", orderID, "error", assert.AnError)

	// Act
	err := useCase.Execute(ctx, orderID, items)

	// Assert
	assert.Error(t, err)
}
