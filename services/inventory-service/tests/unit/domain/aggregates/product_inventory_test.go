package aggregates_test

import (
	"testing"

	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/services/inventory-service/internal/domain/aggregates"
	"ecommerce-platform/services/inventory-service/internal/domain/entities"

	"github.com/stretchr/testify/assert"
)

func TestProductInventory_NewProductInventory(t *testing.T) {
	// Arrange
	currency, _ := valueobjects.NewCurrency("USD")
	price := valueobjects.NewMoney(1000, currency)
	product := entities.NewProduct("product-1", "Test Product", price, "https://example.com/image.jpg")
	stock := entities.NewStock("product-1", 50, 10)

	// Act
	aggregate := aggregates.NewProductInventory(product, stock)

	// Assert
	assert.NotNil(t, aggregate)
	assert.Equal(t, product, aggregate.Product)
	assert.Equal(t, stock, aggregate.Stock)
}

func TestProductInventory_NewProductInventory_WithNilStock(t *testing.T) {
	// Arrange
	currency, _ := valueobjects.NewCurrency("USD")
	price := valueobjects.NewMoney(1000, currency)
	product := entities.NewProduct("product-1", "Test Product", price, "https://example.com/image.jpg")

	// Act
	aggregate := aggregates.NewProductInventory(product, nil)

	// Assert
	assert.NotNil(t, aggregate)
	assert.Equal(t, product, aggregate.Product)
	assert.Nil(t, aggregate.Stock)
}

func TestProductInventory_Validate_Success(t *testing.T) {
	// Arrange
	currency, _ := valueobjects.NewCurrency("USD")
	price := valueobjects.NewMoney(1000, currency)
	product := entities.NewProduct("product-1", "Test Product", price, "https://example.com/image.jpg")
	stock := entities.NewStock("product-1", 50, 10)
	aggregate := aggregates.NewProductInventory(product, stock)

	// Act
	err := aggregate.Validate()

	// Assert
	assert.NoError(t, err)
}

func TestProductInventory_Validate_WithNilStock(t *testing.T) {
	// Arrange
	currency, _ := valueobjects.NewCurrency("USD")
	price := valueobjects.NewMoney(1000, currency)
	product := entities.NewProduct("product-1", "Test Product", price, "https://example.com/image.jpg")
	aggregate := aggregates.NewProductInventory(product, nil)

	// Act
	err := aggregate.Validate()

	// Assert
	assert.NoError(t, err)
}

func TestProductInventory_Validate_InvalidProduct(t *testing.T) {
	// Arrange
	currency, _ := valueobjects.NewCurrency("USD")
	price := valueobjects.NewMoney(1000, currency)
	product := entities.NewProduct("", "Test Product", price, "https://example.com/image.jpg") // Empty ID
	stock := entities.NewStock("product-1", 50, 10)
	aggregate := aggregates.NewProductInventory(product, stock)

	// Act
	err := aggregate.Validate()

	// Assert
	assert.Error(t, err)
}

func TestProductInventory_Validate_InvalidStock(t *testing.T) {
	// Arrange
	currency, _ := valueobjects.NewCurrency("USD")
	price := valueobjects.NewMoney(1000, currency)
	product := entities.NewProduct("product-1", "Test Product", price, "https://example.com/image.jpg")
	stock := entities.NewStock("", 50, 10) // Empty ProductID
	aggregate := aggregates.NewProductInventory(product, stock)

	// Act
	err := aggregate.Validate()

	// Assert
	assert.Error(t, err)
}

func TestProductInventory_ToProduct(t *testing.T) {
	// Arrange
	currency, _ := valueobjects.NewCurrency("USD")
	price := valueobjects.NewMoney(1000, currency)
	product := entities.NewProduct("product-1", "Test Product", price, "https://example.com/image.jpg")
	stock := entities.NewStock("product-1", 50, 10)
	aggregate := aggregates.NewProductInventory(product, stock)

	// Act
	result := aggregate.ToProduct()

	// Assert
	assert.Equal(t, product, result)
}

func TestProductInventory_ToStock(t *testing.T) {
	// Arrange
	currency, _ := valueobjects.NewCurrency("USD")
	price := valueobjects.NewMoney(1000, currency)
	product := entities.NewProduct("product-1", "Test Product", price, "https://example.com/image.jpg")
	stock := entities.NewStock("product-1", 50, 10)
	aggregate := aggregates.NewProductInventory(product, stock)

	// Act
	result := aggregate.ToStock()

	// Assert
	assert.Equal(t, stock, result)
}

func TestProductInventory_ToStock_WithNilStock(t *testing.T) {
	// Arrange
	currency, _ := valueobjects.NewCurrency("USD")
	price := valueobjects.NewMoney(1000, currency)
	product := entities.NewProduct("product-1", "Test Product", price, "https://example.com/image.jpg")
	aggregate := aggregates.NewProductInventory(product, nil)

	// Act
	result := aggregate.ToStock()

	// Assert
	assert.Nil(t, result)
}
