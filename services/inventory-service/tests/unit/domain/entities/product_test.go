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
