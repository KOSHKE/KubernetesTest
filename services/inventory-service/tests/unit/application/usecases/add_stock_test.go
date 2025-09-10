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

func TestAddStockUseCase_Execute_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockInventoryRepositoryFacade(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	useCase := usecases.NewAddStockUseCase(mockRepo, mockLogger)

	ctx := context.Background()
	productID := "product-123"
	quantity := int32(50)

	// Setup mocks
	currency, _ := valueobjects.NewCurrency("USD")
	money := valueobjects.NewMoney(1000, currency)
	product := entities.NewProduct(productID, "Test Product", money, "image.jpg")
	existingStock := entities.NewStock(productID, 100, 0) // Already has 100 units

	mockRepo.EXPECT().GetProductByID(ctx, productID).Return(product, nil)
	mockRepo.EXPECT().GetStockByProductID(ctx, productID).Return(existingStock, nil)
	mockRepo.EXPECT().UpsertStock(ctx, gomock.Any()).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, productID, quantity)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int32(150), result.AvailableQuantity) // 100 + 50
	assert.Equal(t, int32(0), result.ReservedQuantity)
}
