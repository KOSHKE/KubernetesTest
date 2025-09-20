package usecases_test

import (
	"context"
	"testing"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/services/order-service/internal/application/usecases"
	orderValueObjects "ecommerce-platform/services/order-service/internal/domain/valueobjects"
	"ecommerce-platform/services/order-service/tests/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// setupCreateOrderTest creates common test setup for CreateOrderUseCase tests
func setupCreateOrderTest(t *testing.T) (*gomock.Controller, *mocks.MockOrderRepositoryFacade, *usecases.CreateOrderUseCase) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockOrderRepositoryFacade(ctrl)
	useCase := usecases.NewCreateOrderUseCase()
	return ctrl, mockRepo, useCase
}

func TestCreateOrderUseCase_Execute(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockRepo, useCase := setupCreateOrderTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		userID := "user-456"
		shippingAddress := "123 Main St, New York, NY 10001, USA"
		currency := "USD"

		// Create test items
		currency1, _ := valueobjects.NewCurrency("USD")
		price := valueobjects.NewMoney(1000, currency1)
		item, _ := orderValueObjects.NewOrderItem("product-123", "Test Product", 2, price)
		items := []*orderValueObjects.OrderItem{item}

		// Setup mocks
		mockRepo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(nil).
			Times(1)

		// Act
		result, err := useCase.Execute(ctx, userID, shippingAddress, currency, items, mockRepo)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userID, result.UserID)
		assert.Equal(t, shippingAddress, result.ShippingAddress.String())
		assert.Equal(t, currency1, result.Currency)
		assert.Len(t, result.Items, 1)
	})

	t.Run("invalid currency", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockRepo, useCase := setupCreateOrderTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		userID := "user-456"
		shippingAddress := "123 Main St, New York, NY 10001, USA"
		currency := "INVALID"
		items := []*orderValueObjects.OrderItem{}

		// Act
		result, err := useCase.Execute(ctx, userID, shippingAddress, currency, items, mockRepo)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, errors.ErrOrderValidationFailed, err)
		assert.Nil(t, result)
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()
		// Arrange
		ctrl, mockRepo, useCase := setupCreateOrderTest(t)
		defer ctrl.Finish()

		ctx := context.Background()
		userID := "user-456"
		shippingAddress := "123 Main St, New York, NY 10001, USA"
		currency := "USD"
		items := []*orderValueObjects.OrderItem{}

		// Setup mocks
		mockRepo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(assert.AnError).
			Times(1)

		// Act
		result, err := useCase.Execute(ctx, userID, shippingAddress, currency, items, mockRepo)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, errors.ErrOrderPersistenceFailed, err)
		assert.Nil(t, result)
	})
}
