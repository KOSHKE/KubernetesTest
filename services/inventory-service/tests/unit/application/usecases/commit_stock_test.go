package usecases_test

import (
	"context"
	"testing"

	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/services/inventory-service/internal/application/usecases"
	"ecommerce-platform/services/inventory-service/internal/domain/entities"
	"ecommerce-platform/services/inventory-service/tests/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// setupCommitStockTest creates common test setup for CommitStockUseCase tests
func setupCommitStockTest(t *testing.T) (*gomock.Controller, *mocks.MockInventoryRepositoryFacade, *usecases.CommitStockUseCase) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockInventoryRepositoryFacade(ctrl)
	useCase := usecases.NewCommitStockUseCase()
	return ctrl, mockRepo, useCase
}

func TestCommitStockUseCase_Execute(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockRepo, useCase := setupCommitStockTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		orderID := "order-123"
		items := []valueobjects.Item{
			{ProductID: "product-1", Quantity: 10},
			{ProductID: "product-2", Quantity: 5},
		}

		// Setup mocks
		stocks := map[string]*entities.Stock{
			"product-1": entities.NewStock("product-1", 40, 10), // 40 available, 10 reserved
			"product-2": entities.NewStock("product-2", 25, 5),  // 25 available, 5 reserved
		}

		mockRepo.EXPECT().
			GetStocksByProductIDs(ctx, []string{"product-1", "product-2"}, true).
			Return(stocks, nil).
			Times(1)

		mockRepo.EXPECT().
			UpsertStocks(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, stocks []*entities.Stock) error {
				// Verify that stocks were updated with committed quantities
				assert.Len(t, stocks, 2)
				return nil
			}).
			Times(1)

		// Act
		err := useCase.Execute(ctx, orderID, items, mockRepo)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("stock not found", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockRepo, useCase := setupCommitStockTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		orderID := "order-123"
		items := []valueobjects.Item{
			{ProductID: "non-existent-product", Quantity: 10},
		}

		// Setup mocks - stock not found
		stocks := map[string]*entities.Stock{}

		mockRepo.EXPECT().
			GetStocksByProductIDs(ctx, []string{"non-existent-product"}, true).
			Return(stocks, nil).
			Times(1)

		// Act
		err := useCase.Execute(ctx, orderID, items, mockRepo)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "product not found")
	})

	t.Run("insufficient reserved stock", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockRepo, useCase := setupCommitStockTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		orderID := "order-123"
		items := []valueobjects.Item{
			{ProductID: "product-1", Quantity: 20}, // Request more than reserved
		}

		// Setup mocks
		stocks := map[string]*entities.Stock{
			"product-1": entities.NewStock("product-1", 40, 10), // Only 10 reserved
		}

		mockRepo.EXPECT().
			GetStocksByProductIDs(ctx, []string{"product-1"}, true).
			Return(stocks, nil).
			Times(1)

		// Act
		err := useCase.Execute(ctx, orderID, items, mockRepo)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient reserved stock")
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockRepo, useCase := setupCommitStockTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		orderID := "order-123"
		items := []valueobjects.Item{
			{ProductID: "product-1", Quantity: 10},
		}

		// Setup mocks
		mockRepo.EXPECT().
			GetStocksByProductIDs(ctx, []string{"product-1"}, true).
			Return(nil, assert.AnError).
			Times(1)

		// Act
		err := useCase.Execute(ctx, orderID, items, mockRepo)

		// Assert
		assert.Error(t, err)
	})

	t.Run("upsert stocks error", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockRepo, useCase := setupCommitStockTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		orderID := "order-123"
		items := []valueobjects.Item{
			{ProductID: "product-1", Quantity: 10},
		}

		// Setup mocks
		stocks := map[string]*entities.Stock{
			"product-1": entities.NewStock("product-1", 40, 10),
		}

		mockRepo.EXPECT().
			GetStocksByProductIDs(ctx, []string{"product-1"}, true).
			Return(stocks, nil).
			Times(1)

		mockRepo.EXPECT().
			UpsertStocks(ctx, gomock.Any()).
			Return(assert.AnError).
			Times(1)

		// Act
		err := useCase.Execute(ctx, orderID, items, mockRepo)

		// Assert
		assert.Error(t, err)
	})

}
