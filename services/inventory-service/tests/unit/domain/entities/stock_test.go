package entities_test

import (
	"testing"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/services/inventory-service/internal/domain/entities"

	"github.com/stretchr/testify/assert"
)

func TestStock_Reserve(t *testing.T) {
	t.Run("successful reservation", func(t *testing.T) {
		t.Parallel()
		// Arrange
		stock := entities.NewStock("prod1", 100, 0)
		quantity := int32(50)

		// Act
		err := stock.Reserve(quantity)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, int32(50), stock.AvailableQuantity)
		assert.Equal(t, int32(50), stock.ReservedQuantity)
	})

	t.Run("insufficient stock", func(t *testing.T) {
		t.Parallel()
		// Arrange
		stock := entities.NewStock("prod1", 100, 0)
		quantity := int32(150)

		// Act
		err := stock.Reserve(quantity)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, errors.ErrInsufficientStock, err)
		assert.Equal(t, int32(100), stock.AvailableQuantity)
		assert.Equal(t, int32(0), stock.ReservedQuantity)
	})

	t.Run("exact amount", func(t *testing.T) {
		t.Parallel()
		// Arrange
		stock := entities.NewStock("prod1", 100, 0)
		quantity := int32(100)

		// Act
		err := stock.Reserve(quantity)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, int32(0), stock.AvailableQuantity)
		assert.Equal(t, int32(100), stock.ReservedQuantity)
	})
}

func TestStock_Release(t *testing.T) {
	t.Run("successful release", func(t *testing.T) {
		t.Parallel()
		// Arrange
		stock := entities.NewStock("prod1", 50, 50)
		quantity := int32(30)

		// Act
		err := stock.Release(quantity)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, int32(80), stock.AvailableQuantity)
		assert.Equal(t, int32(20), stock.ReservedQuantity)
	})

	t.Run("insufficient reserved stock", func(t *testing.T) {
		t.Parallel()
		// Arrange
		stock := entities.NewStock("prod1", 50, 50)
		quantity := int32(100)

		// Act
		err := stock.Release(quantity)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, errors.ErrInsufficientReservedStock, err)
		assert.Equal(t, int32(50), stock.AvailableQuantity)
		assert.Equal(t, int32(50), stock.ReservedQuantity)
	})

	t.Run("exact reserved amount", func(t *testing.T) {
		t.Parallel()
		// Arrange
		stock := entities.NewStock("prod1", 50, 50)
		quantity := int32(50)

		// Act
		err := stock.Release(quantity)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, int32(100), stock.AvailableQuantity)
		assert.Equal(t, int32(0), stock.ReservedQuantity)
	})
}

func TestStock_Commit(t *testing.T) {
	t.Run("successful commit", func(t *testing.T) {
		t.Parallel()
		// Arrange
		stock := entities.NewStock("prod1", 50, 50)
		quantity := int32(30)

		// Act
		err := stock.Commit(quantity)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, int32(50), stock.AvailableQuantity)
		assert.Equal(t, int32(20), stock.ReservedQuantity)
	})

	t.Run("insufficient reserved stock", func(t *testing.T) {
		t.Parallel()
		// Arrange
		stock := entities.NewStock("prod1", 50, 50)
		quantity := int32(100)

		// Act
		err := stock.Commit(quantity)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, errors.ErrInsufficientReservedStock, err)
		assert.Equal(t, int32(50), stock.AvailableQuantity)
		assert.Equal(t, int32(50), stock.ReservedQuantity)
	})

	t.Run("exact reserved amount", func(t *testing.T) {
		t.Parallel()
		// Arrange
		stock := entities.NewStock("prod1", 50, 50)
		quantity := int32(50)

		// Act
		err := stock.Commit(quantity)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, int32(50), stock.AvailableQuantity)
		assert.Equal(t, int32(0), stock.ReservedQuantity)
	})
}

func TestStock_Validate(t *testing.T) {
	t.Run("valid stock", func(t *testing.T) {
		t.Parallel()
		// Arrange
		stock := entities.NewStock("prod1", 100, 50)

		// Act
		err := stock.Validate()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("empty product ID", func(t *testing.T) {
		t.Parallel()
		// Arrange
		stock := entities.NewStock("", 100, 50)

		// Act
		err := stock.Validate()

		// Assert
		assert.Error(t, err)
		assert.Equal(t, errors.ErrInvalidProductID, err)
	})

	t.Run("negative quantity", func(t *testing.T) {
		t.Parallel()
		// Arrange
		stock := &entities.Stock{ProductID: "prod1", AvailableQuantity: -10, ReservedQuantity: 50}

		// Act
		err := stock.Validate()

		// Assert
		assert.Error(t, err)
		assert.Equal(t, errors.ErrInvalidQuantity, err)
	})
}

func TestStock_GetTotalQuantity(t *testing.T) {
	t.Run("normal stock", func(t *testing.T) {
		t.Parallel()
		// Arrange
		stock := entities.NewStock("prod1", 100, 50)

		// Act
		total := stock.GetTotalQuantity()

		// Assert
		assert.Equal(t, int32(150), total)
	})

	t.Run("zero stock", func(t *testing.T) {
		t.Parallel()
		// Arrange
		stock := entities.NewStock("prod1", 0, 0)

		// Act
		total := stock.GetTotalQuantity()

		// Assert
		assert.Equal(t, int32(0), total)
	})

	t.Run("only available", func(t *testing.T) {
		t.Parallel()
		// Arrange
		stock := entities.NewStock("prod1", 100, 0)

		// Act
		total := stock.GetTotalQuantity()

		// Assert
		assert.Equal(t, int32(100), total)
	})

	t.Run("only reserved", func(t *testing.T) {
		t.Parallel()
		// Arrange
		stock := entities.NewStock("prod1", 0, 50)

		// Act
		total := stock.GetTotalQuantity()

		// Assert
		assert.Equal(t, int32(50), total)
	})
}
