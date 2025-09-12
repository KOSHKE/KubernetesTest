package usecases_test

import (
	"context"
	"testing"

	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/services/inventory-service/internal/application/usecases"
	"ecommerce-platform/services/inventory-service/tests/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCreateProductUseCase_Execute_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockInventoryRepositoryFacade(ctrl)

	useCase := usecases.NewCreateProductUseCase(mockRepo)

	ctx := context.Background()
	name := "Test Product"
	currency, _ := valueobjects.NewCurrency("USD")
	price := valueobjects.NewMoney(1000, currency)
	imageURL := "https://example.com/image.jpg"

	// Setup mocks
	mockRepo.EXPECT().CreateProduct(ctx, gomock.Any()).Return(nil)

	// Act
	result, err := useCase.Execute(ctx, name, price, imageURL)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, name, result.Name)
	assert.Equal(t, price, result.Price)
	assert.Equal(t, imageURL, result.ImageURL)
	assert.NotEmpty(t, result.ID)
	assert.False(t, result.CreatedAt.IsZero())
	assert.False(t, result.UpdatedAt.IsZero())
}

func TestCreateProductUseCase_Execute_RepositoryError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockInventoryRepositoryFacade(ctrl)

	useCase := usecases.NewCreateProductUseCase(mockRepo)

	ctx := context.Background()
	name := "Test Product"
	currency, _ := valueobjects.NewCurrency("USD")
	price := valueobjects.NewMoney(1000, currency)
	imageURL := "https://example.com/image.jpg"

	// Setup mocks
	mockRepo.EXPECT().CreateProduct(ctx, gomock.Any()).Return(assert.AnError)

	// Act
	result, err := useCase.Execute(ctx, name, price, imageURL)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
}
