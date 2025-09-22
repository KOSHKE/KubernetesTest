package usecases_test

import (
	"context"
	"testing"

	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/services/payment-service/internal/application/usecases"
	paymentvalueobjects "ecommerce-platform/services/payment-service/internal/domain/valueobjects"

	"github.com/stretchr/testify/assert"
)

// setupProcessPaymentTest creates common test setup for ProcessPaymentUseCase tests
func setupProcessPaymentTest(t *testing.T) *usecases.ProcessPaymentUseCase {
	useCase := usecases.NewProcessPaymentUseCase()
	return useCase
}

func TestProcessPaymentUseCase_Execute(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		// Arrange
		useCase := setupProcessPaymentTest(t)

		ctx := context.Background()
		orderID := "order-123"
		userID := "user-456"
		currency, _ := valueobjects.NewCurrency("USD")
		amount := valueobjects.NewMoney(1000, currency)
		method := paymentvalueobjects.PaymentMethodCreditCard

		// Act
		payment, err := useCase.Execute(ctx, orderID, userID, amount, method)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, payment)
		assert.Equal(t, orderID, payment.OrderID)
		assert.Equal(t, userID, payment.UserID)
		assert.Equal(t, amount, payment.Amount)
		assert.Equal(t, method, payment.Method)
		assert.NotEmpty(t, payment.ID)
		assert.False(t, payment.CreatedAt.IsZero())
		assert.False(t, payment.UpdatedAt.IsZero())

		// Payment status should be either completed or failed (due to random simulation)
		assert.True(t, payment.Status == paymentvalueobjects.PaymentStatusCompleted ||
			payment.Status == paymentvalueobjects.PaymentStatusFailed)

		// If completed, should have transaction ID
		if payment.Status == paymentvalueobjects.PaymentStatusCompleted {
			assert.NotEmpty(t, payment.TransactionID)
		} else {
			assert.Empty(t, payment.TransactionID)
		}
	})

	t.Run("success with different currency", func(t *testing.T) {
		t.Parallel()
		// Arrange
		useCase := setupProcessPaymentTest(t)

		ctx := context.Background()
		orderID := "order-789"
		userID := "user-101"
		currency, _ := valueobjects.NewCurrency("EUR")
		amount := valueobjects.NewMoney(5000, currency)
		method := paymentvalueobjects.PaymentMethodCreditCard

		// Act
		payment, err := useCase.Execute(ctx, orderID, userID, amount, method)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, payment)
		assert.Equal(t, orderID, payment.OrderID)
		assert.Equal(t, userID, payment.UserID)
		assert.Equal(t, amount, payment.Amount)
		assert.Equal(t, method, payment.Method)
		assert.Equal(t, currency, payment.Amount.Currency)
	})

	t.Run("success with different payment method", func(t *testing.T) {
		t.Parallel()
		// Arrange
		useCase := setupProcessPaymentTest(t)

		ctx := context.Background()
		orderID := "order-456"
		userID := "user-789"
		currency, _ := valueobjects.NewCurrency("USD")
		amount := valueobjects.NewMoney(2500, currency)
		method := paymentvalueobjects.PaymentMethodCreditCard

		// Act
		payment, err := useCase.Execute(ctx, orderID, userID, amount, method)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, payment)
		assert.Equal(t, method, payment.Method)
	})

	t.Run("handles empty order ID", func(t *testing.T) {
		t.Parallel()
		// Arrange
		useCase := setupProcessPaymentTest(t)

		ctx := context.Background()
		orderID := ""
		userID := "user-123"
		currency, _ := valueobjects.NewCurrency("USD")
		amount := valueobjects.NewMoney(1000, currency)
		method := paymentvalueobjects.PaymentMethodCreditCard

		// Act
		payment, err := useCase.Execute(ctx, orderID, userID, amount, method)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, payment)
		assert.Equal(t, orderID, payment.OrderID)
	})

	t.Run("handles empty user ID", func(t *testing.T) {
		t.Parallel()
		// Arrange
		useCase := setupProcessPaymentTest(t)

		ctx := context.Background()
		orderID := "order-123"
		userID := ""
		currency, _ := valueobjects.NewCurrency("USD")
		amount := valueobjects.NewMoney(1000, currency)
		method := paymentvalueobjects.PaymentMethodCreditCard

		// Act
		payment, err := useCase.Execute(ctx, orderID, userID, amount, method)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, payment)
		assert.Equal(t, userID, payment.UserID)
	})

	t.Run("handles zero amount", func(t *testing.T) {
		t.Parallel()
		// Arrange
		useCase := setupProcessPaymentTest(t)

		ctx := context.Background()
		orderID := "order-123"
		userID := "user-456"
		currency, _ := valueobjects.NewCurrency("USD")
		amount := valueobjects.NewMoney(0, currency)
		method := paymentvalueobjects.PaymentMethodCreditCard

		// Act
		payment, err := useCase.Execute(ctx, orderID, userID, amount, method)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, payment)
		assert.Equal(t, amount, payment.Amount)
	})

	t.Run("handles negative amount", func(t *testing.T) {
		t.Parallel()
		// Arrange
		useCase := setupProcessPaymentTest(t)

		ctx := context.Background()
		orderID := "order-123"
		userID := "user-456"
		currency, _ := valueobjects.NewCurrency("USD")
		amount := valueobjects.NewMoney(-100, currency)
		method := paymentvalueobjects.PaymentMethodCreditCard

		// Act
		payment, err := useCase.Execute(ctx, orderID, userID, amount, method)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, payment)
		assert.Equal(t, amount, payment.Amount)
	})
}
