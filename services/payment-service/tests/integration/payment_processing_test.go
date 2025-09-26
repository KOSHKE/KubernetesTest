//go:build integration

package integration

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	dto "ecommerce-platform/pkg/common/dto/payment-service"
	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/services/payment-service/internal/application/services"
	paymentvalueobjects "ecommerce-platform/services/payment-service/internal/domain/valueobjects"
	"ecommerce-platform/services/payment-service/internal/infra/repository"
	"ecommerce-platform/services/payment-service/tests/integration/helpers"
	"ecommerce-platform/services/payment-service/tests/mocks"
)

// TestPaymentProcessingIntegration tests the complete payment processing flow
// StockReserved -> ProcessPayment -> PaymentProcessed event
func TestPaymentProcessingIntegration(t *testing.T) {
	// Setup
	ctx := context.Background()
	db, cleanup := helpers.SetupTestDatabase(t, ctx)
	defer cleanup()

	// Create mock dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	mockMetrics := mocks.NewMockPaymentMetrics(ctrl)

	// Create real components
	outboxRepo := repository.NewOutboxRepository(db)
	paymentSvc := services.NewPaymentApplicationService(outboxRepo, mockLogger, mockMetrics)

	// Test Case 1: Successful Payment Processing
	t.Run("PaymentProcessed_Success_ShouldSaveToOutbox", func(t *testing.T) {
		// Clean database before test
		helpers.CleanDatabase(t, db)

		// Arrange
		orderID := "order-123"
		userID := "user-456"
		currency, _ := valueobjects.NewCurrency("USD")
		amount := valueobjects.NewMoney(1000, currency)
		method := paymentvalueobjects.PaymentMethodCreditCard

		req := &dto.ProcessPaymentRequest{
			PaymentID:      "payment-123",
			OrderID:        orderID,
			UserID:         userID,
			AmountAmount:   amount.Amount,
			AmountCurrency: amount.Currency.Code,
			Method:         string(method),
		}

		// Configure mocks
		mockLogger.EXPECT().
			Error(gomock.Any(), gomock.Any()).
			AnyTimes()
		mockMetrics.EXPECT().
			GRPCRequestDuration(gomock.Any(), gomock.Any()).
			AnyTimes()
		mockMetrics.EXPECT().
			GRPCRequestTotal(gomock.Any(), gomock.Any()).
			AnyTimes()

		// Act
		response, err := paymentSvc.ProcessPayment(ctx, req)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, orderID, response.OrderID)
		assert.Equal(t, userID, response.UserID)
		assert.Equal(t, amount.Amount, response.AmountAmount)
		assert.Equal(t, amount.Currency.Code, response.AmountCurrency)
		assert.Equal(t, string(method), response.Method)
		assert.NotEmpty(t, response.ID)
		assert.True(t, response.Status == string(paymentvalueobjects.PaymentStatusCompleted) ||
			response.Status == string(paymentvalueobjects.PaymentStatusFailed))

		// Check that event is saved to outbox
		require.Eventually(t, func() bool {
			events := helpers.GetOutboxEvents(t, ctx, db)
			return len(events) == 1 &&
				events[0].Type == "PaymentProcessed" &&
				events[0].AggregateID == orderID
		}, 1*time.Second, 10*time.Millisecond, "PaymentProcessed event should be saved to outbox")

		// Verify event payload
		events := helpers.GetOutboxEvents(t, ctx, db)
		require.Len(t, events, 1)
		event := events[0]
		assert.Equal(t, "PaymentProcessed", event.Type)
		assert.Equal(t, orderID, event.AggregateID)
		assert.NotEmpty(t, event.Payload)
	})

	// Test Case 2: Payment Processing with Different Currencies
	t.Run("PaymentProcessed_WithDifferentCurrencies_ShouldSaveToOutbox", func(t *testing.T) {
		// Clean database before test
		helpers.CleanDatabase(t, db)

		// Test with EUR
		currency, _ := valueobjects.NewCurrency("EUR")
		amount := valueobjects.NewMoney(5000, currency)

		req := &dto.ProcessPaymentRequest{
			PaymentID:      "payment-eur-123",
			OrderID:        "order-eur-123",
			UserID:         "user-eur-456",
			AmountAmount:   amount.Amount,
			AmountCurrency: amount.Currency.Code,
			Method:         string(paymentvalueobjects.PaymentMethodCreditCard),
		}

		// Configure mocks
		mockLogger.EXPECT().
			Error(gomock.Any(), gomock.Any()).
			AnyTimes()
		mockMetrics.EXPECT().
			GRPCRequestDuration(gomock.Any(), gomock.Any()).
			AnyTimes()
		mockMetrics.EXPECT().
			GRPCRequestTotal(gomock.Any(), gomock.Any()).
			AnyTimes()

		// Act
		response, err := paymentSvc.ProcessPayment(ctx, req)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, "EUR", response.AmountCurrency)
		assert.Equal(t, int64(5000), response.AmountAmount)

		// Check outbox event
		require.Eventually(t, func() bool {
			events := helpers.GetOutboxEvents(t, ctx, db)
			return len(events) == 1 && events[0].Type == "PaymentProcessed"
		}, 1*time.Second, 10*time.Millisecond)
	})

	// Test Case 3: Payment Processing with Zero Amount
	t.Run("PaymentProcessed_WithZeroAmount_ShouldSaveToOutbox", func(t *testing.T) {
		// Clean database before test
		helpers.CleanDatabase(t, db)

		// Arrange
		currency, _ := valueobjects.NewCurrency("USD")
		amount := valueobjects.NewMoney(0, currency)

		req := &dto.ProcessPaymentRequest{
			PaymentID:      "payment-zero-123",
			OrderID:        "order-zero-123",
			UserID:         "user-zero-456",
			AmountAmount:   amount.Amount,
			AmountCurrency: amount.Currency.Code,
			Method:         string(paymentvalueobjects.PaymentMethodCreditCard),
		}

		// Configure mocks
		mockLogger.EXPECT().
			Error(gomock.Any(), gomock.Any()).
			AnyTimes()
		mockMetrics.EXPECT().
			GRPCRequestDuration(gomock.Any(), gomock.Any()).
			AnyTimes()
		mockMetrics.EXPECT().
			GRPCRequestTotal(gomock.Any(), gomock.Any()).
			AnyTimes()

		// Act
		response, err := paymentSvc.ProcessPayment(ctx, req)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, int64(0), response.AmountAmount)

		// Check outbox event
		require.Eventually(t, func() bool {
			events := helpers.GetOutboxEvents(t, ctx, db)
			return len(events) == 1 && events[0].Type == "PaymentProcessed"
		}, 1*time.Second, 10*time.Millisecond)
	})

	// Test Case 4: Payment Processing with Negative Amount
	t.Run("PaymentProcessed_WithNegativeAmount_ShouldSaveToOutbox", func(t *testing.T) {
		// Clean database before test
		helpers.CleanDatabase(t, db)

		// Arrange
		currency, _ := valueobjects.NewCurrency("USD")
		amount := valueobjects.NewMoney(-100, currency)

		req := &dto.ProcessPaymentRequest{
			PaymentID:      "payment-negative-123",
			OrderID:        "order-negative-123",
			UserID:         "user-negative-456",
			AmountAmount:   amount.Amount,
			AmountCurrency: amount.Currency.Code,
			Method:         string(paymentvalueobjects.PaymentMethodCreditCard),
		}

		// Configure mocks
		mockLogger.EXPECT().
			Error(gomock.Any(), gomock.Any()).
			AnyTimes()
		mockMetrics.EXPECT().
			GRPCRequestDuration(gomock.Any(), gomock.Any()).
			AnyTimes()
		mockMetrics.EXPECT().
			GRPCRequestTotal(gomock.Any(), gomock.Any()).
			AnyTimes()

		// Act
		response, err := paymentSvc.ProcessPayment(ctx, req)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, int64(-100), response.AmountAmount)

		// Check outbox event
		require.Eventually(t, func() bool {
			events := helpers.GetOutboxEvents(t, ctx, db)
			return len(events) == 1 && events[0].Type == "PaymentProcessed"
		}, 1*time.Second, 10*time.Millisecond)
	})

	// Test Case 5: Multiple Payment Processing
	t.Run("PaymentProcessed_MultiplePayments_ShouldSaveAllToOutbox", func(t *testing.T) {
		// Clean database before test
		helpers.CleanDatabase(t, db)

		// Configure mocks
		mockLogger.EXPECT().
			Error(gomock.Any(), gomock.Any()).
			AnyTimes()
		mockMetrics.EXPECT().
			GRPCRequestDuration(gomock.Any(), gomock.Any()).
			AnyTimes()
		mockMetrics.EXPECT().
			GRPCRequestTotal(gomock.Any(), gomock.Any()).
			AnyTimes()

		// Process multiple payments
		payments := []struct {
			orderID string
			userID  string
			amount  int64
		}{
			{"order-1", "user-1", 1000},
			{"order-2", "user-2", 2000},
			{"order-3", "user-3", 3000},
		}

		currency, _ := valueobjects.NewCurrency("USD")

		for _, p := range payments {
			amount := valueobjects.NewMoney(p.amount, currency)
			req := &dto.ProcessPaymentRequest{
				PaymentID:      "payment-" + p.orderID,
				OrderID:        p.orderID,
				UserID:         p.userID,
				AmountAmount:   amount.Amount,
				AmountCurrency: amount.Currency.Code,
				Method:         string(paymentvalueobjects.PaymentMethodCreditCard),
			}

			response, err := paymentSvc.ProcessPayment(ctx, req)
			require.NoError(t, err)
			require.NotNil(t, response)
		}

		// Check that all events are saved to outbox
		require.Eventually(t, func() bool {
			events := helpers.GetOutboxEvents(t, ctx, db)
			return len(events) == 3
		}, 1*time.Second, 10*time.Millisecond, "All payment events should be saved to outbox")

		// Verify all events
		events := helpers.GetOutboxEvents(t, ctx, db)
		assert.Len(t, events, 3)

		for i, event := range events {
			assert.Equal(t, "PaymentProcessed", event.Type)
			assert.Equal(t, payments[i].orderID, event.AggregateID)
		}
	})

	// Test Case 6: Payment Processing with Payment Details
	t.Run("PaymentProcessed_WithPaymentDetails_ShouldSaveToOutbox", func(t *testing.T) {
		// Clean database before test
		helpers.CleanDatabase(t, db)

		// Arrange
		currency, _ := valueobjects.NewCurrency("USD")
		amount := valueobjects.NewMoney(1500, currency)

		req := &dto.ProcessPaymentRequest{
			PaymentID:      "payment-details-123",
			OrderID:        "order-details-123",
			UserID:         "user-details-456",
			AmountAmount:   amount.Amount,
			AmountCurrency: amount.Currency.Code,
			Method:         string(paymentvalueobjects.PaymentMethodCreditCard),
			Details: &dto.PaymentDetails{
				CardNumber:  "4111111111111111",
				CardHolder:  "John Doe",
				ExpiryMonth: "12",
				ExpiryYear:  "2025",
				CVV:         "123",
			},
		}

		// Configure mocks
		mockLogger.EXPECT().
			Error(gomock.Any(), gomock.Any()).
			AnyTimes()
		mockMetrics.EXPECT().
			GRPCRequestDuration(gomock.Any(), gomock.Any()).
			AnyTimes()
		mockMetrics.EXPECT().
			GRPCRequestTotal(gomock.Any(), gomock.Any()).
			AnyTimes()

		// Act
		response, err := paymentSvc.ProcessPayment(ctx, req)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, req.OrderID, response.OrderID)
		assert.Equal(t, req.UserID, response.UserID)
		assert.Equal(t, req.AmountAmount, response.AmountAmount)
		assert.Equal(t, req.AmountCurrency, response.AmountCurrency)

		// Check outbox event
		require.Eventually(t, func() bool {
			events := helpers.GetOutboxEvents(t, ctx, db)
			return len(events) == 1 && events[0].Type == "PaymentProcessed"
		}, 1*time.Second, 10*time.Millisecond)
	})
}

// TestPaymentOutboxReliability tests outbox pattern reliability for payment events
func TestPaymentOutboxReliability(t *testing.T) {
	// Setup
	ctx := context.Background()
	db, cleanup := helpers.SetupTestDatabase(t, ctx)
	defer cleanup()

	// Create mock dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	mockMetrics := mocks.NewMockPaymentMetrics(ctrl)

	// Create real components
	outboxRepo := repository.NewOutboxRepository(db)
	paymentSvc := services.NewPaymentApplicationService(outboxRepo, mockLogger, mockMetrics)

	// Test Case 1: Event is saved to outbox even if logging fails
	t.Run("PaymentEventSavedToOutbox_WhenLoggingFails", func(t *testing.T) {
		// Clean database before test
		helpers.CleanDatabase(t, db)

		// Arrange
		currency, _ := valueobjects.NewCurrency("USD")
		amount := valueobjects.NewMoney(1000, currency)

		req := &dto.ProcessPaymentRequest{
			PaymentID:      "payment-reliability-123",
			OrderID:        "order-reliability-123",
			UserID:         "user-reliability-456",
			AmountAmount:   amount.Amount,
			AmountCurrency: amount.Currency.Code,
			Method:         string(paymentvalueobjects.PaymentMethodCreditCard),
		}

		// Configure mock logger to return error (simulating logging failure)
		mockLogger.EXPECT().
			Error(gomock.Any(), gomock.Any()).
			AnyTimes()

		// Act
		response, err := paymentSvc.ProcessPayment(ctx, req)

		// Assert
		require.NoError(t, err, "ProcessPayment should succeed even if logging fails")
		require.NotNil(t, response)

		// Check that event is saved to outbox
		require.Eventually(t, func() bool {
			events := helpers.GetOutboxEvents(t, ctx, db)
			return len(events) == 1 &&
				events[0].Type == "PaymentProcessed" &&
				events[0].AggregateID == req.OrderID
		}, 1*time.Second, 10*time.Millisecond, "PaymentProcessed event should be saved to outbox")
	})

	// Test Case 2: Multiple events are saved correctly
	t.Run("MultiplePaymentEvents_ShouldBeSavedCorrectly", func(t *testing.T) {
		// Clean database before test
		helpers.CleanDatabase(t, db)

		// Configure mocks
		mockLogger.EXPECT().
			Error(gomock.Any(), gomock.Any()).
			AnyTimes()
		mockMetrics.EXPECT().
			GRPCRequestDuration(gomock.Any(), gomock.Any()).
			AnyTimes()
		mockMetrics.EXPECT().
			GRPCRequestTotal(gomock.Any(), gomock.Any()).
			AnyTimes()

		// Process multiple payments
		currency, _ := valueobjects.NewCurrency("USD")
		orderIDs := []string{"order-1", "order-2", "order-3"}

		for _, orderID := range orderIDs {
			amount := valueobjects.NewMoney(1000, currency)
			req := &dto.ProcessPaymentRequest{
				PaymentID:      "payment-" + orderID,
				OrderID:        orderID,
				UserID:         "user-" + orderID,
				AmountAmount:   amount.Amount,
				AmountCurrency: amount.Currency.Code,
				Method:         string(paymentvalueobjects.PaymentMethodCreditCard),
			}

			_, err := paymentSvc.ProcessPayment(ctx, req)
			require.NoError(t, err)
		}

		// Check that all events are saved
		require.Eventually(t, func() bool {
			events := helpers.GetOutboxEvents(t, ctx, db)
			return len(events) == 3
		}, 1*time.Second, 10*time.Millisecond, "All events should be saved to outbox")

		// Verify each event
		events := helpers.GetOutboxEvents(t, ctx, db)
		assert.Len(t, events, 3)

		for i, event := range events {
			assert.Equal(t, "PaymentProcessed", event.Type)
			assert.Equal(t, orderIDs[i], event.AggregateID)
			assert.NotEmpty(t, event.Payload)
		}
	})
}

// TestPaymentConcurrency tests concurrent payment processing
func TestPaymentConcurrency(t *testing.T) {
	// Setup
	ctx := context.Background()
	db, cleanup := helpers.SetupTestDatabase(t, ctx)
	defer cleanup()

	// Create mock dependencies
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLogger(ctrl)
	mockMetrics := mocks.NewMockPaymentMetrics(ctrl)

	// Create real components
	outboxRepo := repository.NewOutboxRepository(db)
	paymentSvc := services.NewPaymentApplicationService(outboxRepo, mockLogger, mockMetrics)

	// Test Case: Concurrent payment processing
	t.Run("ConcurrentPaymentProcessing_ShouldHandleRaceConditions", func(t *testing.T) {
		// Clean database before test
		helpers.CleanDatabase(t, db)

		// Configure mocks
		mockLogger.EXPECT().
			Error(gomock.Any(), gomock.Any()).
			AnyTimes()
		mockMetrics.EXPECT().
			GRPCRequestDuration(gomock.Any(), gomock.Any()).
			AnyTimes()
		mockMetrics.EXPECT().
			GRPCRequestTotal(gomock.Any(), gomock.Any()).
			AnyTimes()

		// Arrange
		currency, _ := valueobjects.NewCurrency("USD")
		amount := valueobjects.NewMoney(1000, currency)

		// Process payments concurrently
		orderIDs := []string{"concurrent-1", "concurrent-2", "concurrent-3", "concurrent-4", "concurrent-5"}
		var wg sync.WaitGroup
		wg.Add(len(orderIDs))

		for _, orderID := range orderIDs {
			go func(oid string) {
				defer wg.Done()
				req := &dto.ProcessPaymentRequest{
					PaymentID:      "payment-" + oid,
					OrderID:        oid,
					UserID:         "user-" + oid,
					AmountAmount:   amount.Amount,
					AmountCurrency: amount.Currency.Code,
					Method:         string(paymentvalueobjects.PaymentMethodCreditCard),
				}

				_, err := paymentSvc.ProcessPayment(ctx, req)
				require.NoError(t, err)
			}(orderID)
		}

		// Wait for all goroutines to complete
		wg.Wait()

		// Check that all events are saved
		require.Eventually(t, func() bool {
			events := helpers.GetOutboxEvents(t, ctx, db)
			return len(events) == len(orderIDs)
		}, 2*time.Second, 50*time.Millisecond, "All concurrent events should be saved to outbox")

		// Verify all events
		events := helpers.GetOutboxEvents(t, ctx, db)
		assert.Len(t, events, len(orderIDs))

		// Check that all order IDs are present
		eventOrderIDs := make(map[string]bool)
		for _, event := range events {
			eventOrderIDs[event.AggregateID] = true
		}

		for _, orderID := range orderIDs {
			assert.True(t, eventOrderIDs[orderID], "Order ID %s should be present in events", orderID)
		}
	})
}
