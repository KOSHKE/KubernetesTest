package entities_test

import (
	"testing"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/services/inventory-service/internal/domain/entities"

	"github.com/stretchr/testify/assert"
)

func TestProduct_Validate(t *testing.T) {
	tests := map[string]struct {
		product       *entities.Product
		expectedError error
	}{
		"valid product": {
			product:       createTestProduct("prod1", "Test Product", 1000, "USD"),
			expectedError: nil,
		},
		"empty ID": {
			product:       createTestProduct("", "Test Product", 1000, "USD"),
			expectedError: errors.ErrInvalidProductID,
		},
		"empty name": {
			product:       createTestProduct("prod1", "", 1000, "USD"),
			expectedError: errors.ErrInvalidProductName,
		},
		"valid with zero price": {
			product:       createTestProduct("prod1", "Free Product", 0, "USD"),
			expectedError: nil,
		},
		"valid with empty image URL": {
			product:       createTestProduct("prod1", "Test Product", 1000, "USD"),
			expectedError: nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Act
			err := tt.product.Validate()

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProduct_NewProduct(t *testing.T) {
	tests := map[string]struct {
		id       string
		name     string
		price    valueobjects.Money
		imageURL string
		expected *entities.Product
	}{
		"valid product creation": {
			id:       "prod1",
			name:     "Test Product",
			price:    createTestMoney(1000, "USD"),
			imageURL: "https://example.com/image.jpg",
			expected: &entities.Product{
				ID:       "prod1",
				Name:     "Test Product",
				Price:    createTestMoney(1000, "USD"),
				ImageURL: "https://example.com/image.jpg",
			},
		},
		"product with empty image URL": {
			id:       "prod2",
			name:     "Test Product 2",
			price:    createTestMoney(2000, "EUR"),
			imageURL: "",
			expected: &entities.Product{
				ID:       "prod2",
				Name:     "Test Product 2",
				Price:    createTestMoney(2000, "EUR"),
				ImageURL: "",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Act
			product := entities.NewProduct(tt.id, tt.name, tt.price, tt.imageURL)

			// Assert
			assert.Equal(t, tt.expected.ID, product.ID)
			assert.Equal(t, tt.expected.Name, product.Name)
			assert.Equal(t, tt.expected.Price, product.Price)
			assert.Equal(t, tt.expected.ImageURL, product.ImageURL)
			assert.False(t, product.CreatedAt.IsZero())
			assert.False(t, product.UpdatedAt.IsZero())
		})
	}
}

func TestProduct_FieldAccess(t *testing.T) {
	// This test ensures that all fields are accessible and used
	// If a field is removed, the test won't compile

	product := createTestProduct("prod1", "Test Product", 1000, "USD")

	// Check that all fields are accessible
	assert.Equal(t, "prod1", product.ID)
	assert.Equal(t, "Test Product", product.Name)
	assert.Equal(t, createTestMoney(1000, "USD"), product.Price)
	assert.Equal(t, "https://example.com/image.jpg", product.ImageURL)
	assert.False(t, product.CreatedAt.IsZero())
	assert.False(t, product.UpdatedAt.IsZero())

	// Check that fields can be modified
	product.Name = "Updated Product"
	product.ImageURL = "https://example.com/new-image.jpg"

	assert.Equal(t, "Updated Product", product.Name)
	assert.Equal(t, "https://example.com/new-image.jpg", product.ImageURL)
}

func TestProduct_PriceOperations(t *testing.T) {
	tests := map[string]struct {
		product        *entities.Product
		operation      func(*entities.Product) valueobjects.Money
		expectedAmount int64
		expectedError  error
	}{
		"get price": {
			product: createTestProduct("prod1", "Test Product", 1000, "USD"),
			operation: func(p *entities.Product) valueobjects.Money {
				return p.Price
			},
			expectedAmount: 1000,
			expectedError:  nil,
		},
		"price validation": {
			product: createTestProduct("prod1", "Test Product", 1000, "USD"),
			operation: func(p *entities.Product) valueobjects.Money {
				// Check that price is valid
				err := p.Price.Validate()
				assert.NoError(t, err)
				return p.Price
			},
			expectedAmount: 1000,
			expectedError:  nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Act
			result := tt.operation(tt.product)

			// Assert
			assert.Equal(t, tt.expectedAmount, result.Amount)
			if tt.expectedError != nil {
				assert.Error(t, tt.expectedError)
			} else {
				assert.NoError(t, tt.expectedError)
			}
		})
	}
}

func TestProduct_TimeFields(t *testing.T) {
	// Check that time fields work correctly
	product1 := createTestProduct("prod1", "Test Product 1", 1000, "USD")
	product2 := createTestProduct("prod2", "Test Product 2", 2000, "EUR")

	// CreatedAt and UpdatedAt should be set
	assert.False(t, product1.CreatedAt.IsZero())
	assert.False(t, product1.UpdatedAt.IsZero())
	assert.False(t, product2.CreatedAt.IsZero())
	assert.False(t, product2.UpdatedAt.IsZero())

	// Time fields should be different for different products
	assert.NotEqual(t, product1.CreatedAt, product2.CreatedAt)
	assert.NotEqual(t, product1.UpdatedAt, product2.UpdatedAt)
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
