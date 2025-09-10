package entities_test

import (
	"testing"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/services/inventory-service/internal/domain/entities"

	"github.com/stretchr/testify/assert"
)

func TestStock_Reserve(t *testing.T) {
	tests := map[string]struct {
		stock             *entities.Stock
		quantity          int32
		expectedAvailable int32
		expectedReserved  int32
		expectedError     error
	}{
		"successful reservation": {
			stock:             entities.NewStock("prod1", 100, 0),
			quantity:          50,
			expectedAvailable: 50,
			expectedReserved:  50,
			expectedError:     nil,
		},
		"insufficient stock": {
			stock:             entities.NewStock("prod1", 100, 0),
			quantity:          150,
			expectedAvailable: 100,
			expectedReserved:  0,
			expectedError:     errors.ErrInsufficientStock,
		},
		"exact amount": {
			stock:             entities.NewStock("prod1", 100, 0),
			quantity:          100,
			expectedAvailable: 0,
			expectedReserved:  100,
			expectedError:     nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Act
			err := tt.stock.Reserve(tt.quantity)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedAvailable, tt.stock.AvailableQuantity)
			assert.Equal(t, tt.expectedReserved, tt.stock.ReservedQuantity)
		})
	}
}

func TestStock_Release(t *testing.T) {
	tests := map[string]struct {
		stock             *entities.Stock
		quantity          int32
		expectedAvailable int32
		expectedReserved  int32
		expectedError     error
	}{
		"successful release": {
			stock:             entities.NewStock("prod1", 50, 50),
			quantity:          30,
			expectedAvailable: 80,
			expectedReserved:  20,
			expectedError:     nil,
		},
		"insufficient reserved stock": {
			stock:             entities.NewStock("prod1", 50, 50),
			quantity:          100,
			expectedAvailable: 50,
			expectedReserved:  50,
			expectedError:     errors.ErrInsufficientReservedStock,
		},
		"exact reserved amount": {
			stock:             entities.NewStock("prod1", 50, 50),
			quantity:          50,
			expectedAvailable: 100,
			expectedReserved:  0,
			expectedError:     nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Act
			err := tt.stock.Release(tt.quantity)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedAvailable, tt.stock.AvailableQuantity)
			assert.Equal(t, tt.expectedReserved, tt.stock.ReservedQuantity)
		})
	}
}

func TestStock_Commit(t *testing.T) {
	tests := map[string]struct {
		stock             *entities.Stock
		quantity          int32
		expectedAvailable int32
		expectedReserved  int32
		expectedError     error
	}{
		"successful commit": {
			stock:             entities.NewStock("prod1", 50, 50),
			quantity:          30,
			expectedAvailable: 50,
			expectedReserved:  20,
			expectedError:     nil,
		},
		"insufficient reserved stock": {
			stock:             entities.NewStock("prod1", 50, 50),
			quantity:          100,
			expectedAvailable: 50,
			expectedReserved:  50,
			expectedError:     errors.ErrInsufficientReservedStock,
		},
		"exact reserved amount": {
			stock:             entities.NewStock("prod1", 50, 50),
			quantity:          50,
			expectedAvailable: 50,
			expectedReserved:  0,
			expectedError:     nil,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Act
			err := tt.stock.Commit(tt.quantity)

			// Assert
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedAvailable, tt.stock.AvailableQuantity)
			assert.Equal(t, tt.expectedReserved, tt.stock.ReservedQuantity)
		})
	}
}

func TestStock_Validate(t *testing.T) {
	tests := map[string]struct {
		stock         *entities.Stock
		expectedError error
	}{
		"valid stock": {
			stock:         entities.NewStock("prod1", 100, 50),
			expectedError: nil,
		},
		"empty product ID": {
			stock:         entities.NewStock("", 100, 50),
			expectedError: errors.ErrInvalidProductID,
		},
		"negative quantity": {
			stock:         &entities.Stock{ProductID: "prod1", AvailableQuantity: -10, ReservedQuantity: 50},
			expectedError: errors.ErrInvalidQuantity,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Act
			err := tt.stock.Validate()

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

func TestStock_GetTotalQuantity(t *testing.T) {
	tests := map[string]struct {
		stock         *entities.Stock
		expectedTotal int32
	}{
		"normal stock": {
			stock:         entities.NewStock("prod1", 100, 50),
			expectedTotal: 150,
		},
		"zero stock": {
			stock:         entities.NewStock("prod1", 0, 0),
			expectedTotal: 0,
		},
		"only available": {
			stock:         entities.NewStock("prod1", 100, 0),
			expectedTotal: 100,
		},
		"only reserved": {
			stock:         entities.NewStock("prod1", 0, 50),
			expectedTotal: 50,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Act
			total := tt.stock.GetTotalQuantity()

			// Assert
			assert.Equal(t, tt.expectedTotal, total)
		})
	}
}
