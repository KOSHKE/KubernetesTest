//go:build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/proto-go/events"
	"ecommerce-platform/services/order-service/internal/application/dto"
	"ecommerce-platform/services/order-service/internal/application/services"
	"ecommerce-platform/services/order-service/internal/domain/valueobjects"
	"ecommerce-platform/services/order-service/internal/infra/consumer"
	"ecommerce-platform/services/order-service/internal/infra/repository"
	"ecommerce-platform/services/order-service/tests/integration/helpers"
)

var (
	testDB        *gorm.DB
	testDBCleanup func()
)

// TestMain sets up the test database once for all tests in the package
func TestMain(m *testing.M) {
	ctx := context.Background()

	// Setup database once
	testDB, testDBCleanup = helpers.SetupTestDatabase(&testing.T{}, ctx)
	if testDB == nil {
		panic("failed to setup test database")
	}

	// Run tests
	code := m.Run()

	// Cleanup
	if testDBCleanup != nil {
		testDBCleanup()
	}

	os.Exit(code)
}

// TestOrderLifecycleIntegration tests the complete order lifecycle with event processing
// OrderCreated -> StockReserved -> PaymentProcessed -> StockCommitted/OrderCancelled
func TestOrderLifecycleIntegration(t *testing.T) {
	// Setup
	ctx := context.Background()

	// Create real components using shared test database
	repo := repository.NewOrderRepositoryFacade(testDB)
	log := logger.NewZapLogger(zap.NewNop().Sugar())
	appService := services.NewOrderApplicationService(repo, log)
	eventHandlers := consumer.NewEventHandlers(appService, log)

	// Test Case 1: Complete Happy Path - Order Creation to Completion
	t.Run("CompleteOrderLifecycle_HappyPath", func(t *testing.T) {
		// Clean database before test
		helpers.CleanDatabase(t, testDB)

		// Step 1: Create Order
		orderID := helpers.CreateTestOrder(t, ctx, appService, "user-123", "USD", []dto.OrderItemRequest{
			helpers.CreateOrderItemRequest("product-1", "Test Product", 2, 1000, "USD"),
		})

		// Verify OrderCreated event is saved to outbox
		require.Eventually(t, func() bool {
			outboxEvents := helpers.GetOutboxEvents(t, ctx, testDB)
			return len(outboxEvents) == 1 &&
				outboxEvents[0].Type == "OrderCreated" &&
				outboxEvents[0].AggregateID == orderID
		}, 2*time.Second, 10*time.Millisecond, "OrderCreated event should be saved to outbox")

		// Step 2: Simulate StockReserved event from inventory service
		stockReservedEvent := &events.StockReserved{
			OrderId: orderID,
		}

		err := eventHandlers.HandleStockReserved(ctx, stockReservedEvent)
		require.NoError(t, err)

		// Verify order status is updated to Confirmed
		order := helpers.GetOrderByID(t, ctx, appService, orderID, "user-123")
		require.Equal(t, valueobjects.OrderStatusConfirmed, order.Status)

		// Step 3: Simulate successful PaymentProcessed event from payment service
		paymentProcessedEvent := &events.PaymentProcessed{
			OrderId:  orderID,
			Success:  true,
			Amount:   2000, // 2 * 1000
			Currency: "USD",
		}

		err = eventHandlers.HandlePaymentProcessed(ctx, paymentProcessedEvent)
		require.NoError(t, err)

		// Verify order status is updated to Paid
		order = helpers.GetOrderByID(t, ctx, appService, orderID, "user-123")
		require.Equal(t, valueobjects.OrderStatusPaid, order.Status)

		// Step 4: Simulate StockCommitted event from inventory service
		stockCommittedEvent := &events.StockCommitted{
			OrderId: orderID,
		}

		err = eventHandlers.HandleStockCommitted(ctx, stockCommittedEvent)
		require.NoError(t, err)

		// Verify order status is updated to Completed
		order = helpers.GetOrderByID(t, ctx, appService, orderID, "user-123")
		require.Equal(t, valueobjects.OrderStatusCompleted, order.Status)
		require.Equal(t, int64(2000), order.TotalAmount.Amount)
	})

	// Test Case 2: Payment Failed - Order Should Be Marked as PaymentFailed
	t.Run("PaymentFailed_ShouldUpdateOrderStatus", func(t *testing.T) {
		// Clean database before test
		helpers.CleanDatabase(t, testDB)

		// Create order and simulate stock reservation
		orderID := helpers.CreateTestOrder(t, ctx, appService, "user-456", "USD", []dto.OrderItemRequest{
			helpers.CreateOrderItemRequest("product-2", "Another Product", 1, 500, "USD"),
		})

		// Simulate StockReserved
		stockReservedEvent := &events.StockReserved{
			OrderId: orderID,
		}
		err := eventHandlers.HandleStockReserved(ctx, stockReservedEvent)
		require.NoError(t, err)

		// Simulate failed PaymentProcessed event
		paymentProcessedEvent := &events.PaymentProcessed{
			OrderId:  orderID,
			Success:  false,
			Amount:   500,
			Currency: "USD",
			Message:  "Insufficient funds",
		}

		err = eventHandlers.HandlePaymentProcessed(ctx, paymentProcessedEvent)
		require.NoError(t, err)

		// Verify order status is updated to PaymentFailed
		order := helpers.GetOrderByID(t, ctx, appService, orderID, "user-456")
		require.Equal(t, valueobjects.OrderStatusPaymentFailed, order.Status)
	})

	// Test Case 3: Order Cancellation - Should Publish OrderCancelled Event
	t.Run("OrderCancellation_ShouldPublishEvent", func(t *testing.T) {
		// Clean database before test
		helpers.CleanDatabase(t, testDB)

		// Create order
		orderID := helpers.CreateTestOrder(t, ctx, appService, "user-789", "EUR", []dto.OrderItemRequest{
			helpers.CreateOrderItemRequest("product-3", "Cancellable Product", 3, 750, "EUR"),
		})

		// Cancel the order
		cancelReq := &dto.CancelOrderRequest{
			OrderID: orderID,
			UserID:  "user-789",
			Reason:  "USER_REQUEST",
		}

		_, err := appService.CancelOrder(ctx, cancelReq)
		require.NoError(t, err)

		// Verify order status is Cancelled
		order := helpers.GetOrderByID(t, ctx, appService, orderID, "user-789")
		require.Equal(t, valueobjects.OrderStatusCancelled, order.Status)

		// Verify OrderCancelled event is saved to outbox
		require.Eventually(t, func() bool {
			outboxEvents := helpers.GetOutboxEvents(t, ctx, testDB)
			cancelledEvents := helpers.FilterEventsByType(outboxEvents, "OrderCancelled")
			return len(cancelledEvents) == 1 && cancelledEvents[0].AggregateID == orderID
		}, 2*time.Second, 10*time.Millisecond, "OrderCancelled event should be saved to outbox")
	})

	// Test Case 4: Stock Released After Payment Failure
	t.Run("StockReleased_ShouldUpdateOrderStatus", func(t *testing.T) {
		// Clean database before test
		helpers.CleanDatabase(t, testDB)

		// Create order
		orderID := helpers.CreateTestOrder(t, ctx, appService, "user-stock", "USD", []dto.OrderItemRequest{
			helpers.CreateOrderItemRequest("product-stock", "Stock Product", 1, 1500, "USD"),
		})

		// Simulate StockReleased event (e.g., after payment failure)
		stockReleasedEvent := &events.StockReleased{
			OrderId: orderID,
		}

		err := eventHandlers.HandleStockReleased(ctx, stockReleasedEvent)
		require.NoError(t, err)

		// Verify order status is updated to StockReleased
		order := helpers.GetOrderByID(t, ctx, appService, orderID, "user-stock")
		require.Equal(t, valueobjects.OrderStatusStockReleased, order.Status)
	})

	// Test Case 5: Concurrent Order Processing
	t.Run("ConcurrentOrderProcessing_ShouldHandleRaceConditions", func(t *testing.T) {
		// Clean database before test
		helpers.CleanDatabase(t, testDB)

		// Create multiple orders concurrently using channels to avoid data race
		type orderResult struct {
			orderID string
			userID  string
			err     error
		}

		results := make(chan orderResult, 3)

		for i := 0; i < 3; i++ {
			go func(index int) {
				userID := fmt.Sprintf("concurrent-user-%d", index+1)
				productID := fmt.Sprintf("concurrent-product-%d", index+1)
				orderID := ""
				var err error

				func() {
					defer func() {
						if r := recover(); r != nil {
							switch x := r.(type) {
							case error:
								err = x
							default:
								err = fmt.Errorf("panic: %v", r)
							}
						}
					}()
					orderID = helpers.CreateTestOrder(t, ctx, appService,
						userID,
						"USD",
						[]dto.OrderItemRequest{
							helpers.CreateOrderItemRequest(productID, "Concurrent Product", 1, 100, "USD"),
						})
				}()

				results <- orderResult{orderID: orderID, userID: userID, err: err}
			}(i)
		}

		// Collect results
		var successfulOrders []orderResult
		for i := 0; i < 3; i++ {
			result := <-results
			if result.err == nil && result.orderID != "" {
				successfulOrders = append(successfulOrders, result)
			}
		}

		// Verify all orders were created successfully
		require.Equal(t, 3, len(successfulOrders), "All concurrent orders should be created successfully")

		for _, result := range successfulOrders {
			order := helpers.GetOrderByID(t, ctx, appService, result.orderID, result.userID)
			require.Equal(t, valueobjects.OrderStatusPending, order.Status)
		}
	})

	// Test Case 6: Order Access Control
	t.Run("OrderAccessControl_ShouldPreventUnauthorizedAccess", func(t *testing.T) {
		// Clean database before test
		helpers.CleanDatabase(t, testDB)

		// Create order for user-1
		orderID := helpers.CreateTestOrder(t, ctx, appService, "user-owner", "USD", []dto.OrderItemRequest{
			helpers.CreateOrderItemRequest("access-product", "Access Product", 1, 200, "USD"),
		})

		// Try to cancel order as different user
		cancelReq := &dto.CancelOrderRequest{
			OrderID: orderID,
			UserID:  "user-hacker", // Different user
			Reason:  "UNAUTHORIZED_ATTEMPT",
		}

		_, err := appService.CancelOrder(ctx, cancelReq)
		require.Error(t, err)

		// Verify order status is still Pending
		order := helpers.GetOrderByID(t, ctx, appService, orderID, "user-owner")
		require.Equal(t, valueobjects.OrderStatusPending, order.Status)
	})
}

// TestOutboxPatternReliability tests outbox pattern reliability for order service
func TestOutboxPatternReliability(t *testing.T) {
	// Setup
	ctx := context.Background()

	// Create real components using shared test database
	repo := repository.NewOrderRepositoryFacade(testDB)
	log := logger.NewZapLogger(zap.NewNop().Sugar())
	appService := services.NewOrderApplicationService(repo, log)

	// Test Case 1: Event is saved to outbox
	t.Run("EventSavedToOutbox", func(t *testing.T) {
		// Clean database before test
		helpers.CleanDatabase(t, testDB)

		// Create order
		orderID := helpers.CreateTestOrder(t, ctx, appService, "outbox-user", "USD", []dto.OrderItemRequest{
			helpers.CreateOrderItemRequest("outbox-product", "Outbox Product", 1, 300, "USD"),
		})

		// Check that event is saved to outbox
		require.Eventually(t, func() bool {
			outboxEvents := helpers.GetOutboxEvents(t, ctx, testDB)
			return len(outboxEvents) == 1 &&
				outboxEvents[0].Type == "OrderCreated" &&
				outboxEvents[0].AggregateID == orderID
		}, 2*time.Second, 10*time.Millisecond, "OrderCreated event should be saved to outbox")
	})
}
