package entities_test

import (
	"testing"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/services/inventory-service/internal/domain/entities"

	"github.com/stretchr/testify/assert"
)

func TestProduct_Validate(t *testing.T) {
	t.Run("valid product", func(t *testing.T) {
		t.Parallel()
		// Arrange
		product := createTestProduct("prod1", "Test Product", 1000, "USD")

		// Act
		err := product.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("empty ID", func(t *testing.T) {
		t.Parallel()
		// Arrange
		product := createTestProduct("", "Test Product", 1000, "USD")

		// Act
		err := product.Validate()

		// Assert
		assert.Error(t, err)
		assert.Equal(t, errors.ErrInvalidProductID, err)
	})

	t.Run("empty name", func(t *testing.T) {
		t.Parallel()
		// Arrange
		product := createTestProduct("prod1", "", 1000, "USD")

		// Act
		err := product.Validate()

		// Assert
		assert.Error(t, err)
		assert.Equal(t, errors.ErrInvalidProductName, err)
	})

	t.Run("valid with zero price", func(t *testing.T) {
		t.Parallel()
		// Arrange
		product := createTestProduct("prod1", "Free Product", 0, "USD")

		// Act
		err := product.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("valid with empty image URL", func(t *testing.T) {
		t.Parallel()
		// Arrange
		product := createTestProduct("prod1", "Test Product", 1000, "USD")

		// Act
		err := product.Validate()

		// Assert
		assert.NoError(t, err)
	})
}

func TestProduct_NewProduct(t *testing.T) {
	t.Run("valid product creation", func(t *testing.T) {
		t.Parallel()
		// Arrange
		id := "prod1"
		name := "Test Product"
		price := createTestMoney(1000, "USD")
		imageURL := "https://example.com/image.jpg"

		// Act
		product := entities.NewProduct(id, name, price, imageURL)

		// Assert
		assert.Equal(t, id, product.ID)
		assert.Equal(t, name, product.Name)
		assert.Equal(t, price, product.Price)
		assert.Equal(t, imageURL, product.ImageURL)
		assert.False(t, product.CreatedAt.IsZero())
		assert.False(t, product.UpdatedAt.IsZero())
	})

	t.Run("product with empty image URL", func(t *testing.T) {
		t.Parallel()
		// Arrange
		id := "prod2"
		name := "Test Product 2"
		price := createTestMoney(2000, "EUR")
		imageURL := ""

		// Act
		product := entities.NewProduct(id, name, price, imageURL)

		// Assert
		assert.Equal(t, id, product.ID)
		assert.Equal(t, name, product.Name)
		assert.Equal(t, price, product.Price)
		assert.Equal(t, imageURL, product.ImageURL)
		assert.False(t, product.CreatedAt.IsZero())
		assert.False(t, product.UpdatedAt.IsZero())
	})
}

// Helper functions for creating test data
func createTestProduct(id, name string, amount int64, currency string) *entities.Product {
	currencyObj, _ := valueobjects.NewCurrency(currency)
	money := valueobjects.NewMoney(amount, currencyObj)
	return entities.NewProduct(id, name, money, "https://example.com/image.jpg")
}

func createTestMoney(amount int64, currency string) valueobjects.Money {
	currencyObj, _ := valueobjects.NewCurrency(currency)
	return valueobjects.NewMoney(amount, currencyObj)
}
