//go:build integration

package integration

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"

	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/pkg/outbox"
	"ecommerce-platform/proto-go/common"
	"ecommerce-platform/proto-go/events"
	"ecommerce-platform/services/inventory-service/internal/application/dto"
	"ecommerce-platform/services/inventory-service/internal/application/services"
	"ecommerce-platform/services/inventory-service/internal/infra/consumer"
	"ecommerce-platform/services/inventory-service/internal/infra/migration"
	"ecommerce-platform/services/inventory-service/internal/infra/repository"
	"ecommerce-platform/services/inventory-service/tests/mocks"
)

// TestEventProcessingIntegration tests the complete event processing cycle
// OrderCreated -> StockReserved -> PaymentProcessed -> StockCommitted/StockReleased
func TestEventProcessingIntegration(t *testing.T) {
	// Setup
	ctx := context.Background()
	db, cleanup := setupTestDatabase(t, ctx)
	defer cleanup()

	// Create mock publisher for event publishing verification
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPublisher := mocks.NewMockEventPublisher(ctrl)

	// Create real components
	repo := repository.NewInventoryRepository(db)
	// Create a simple logger for tests
	// Use silent logger for tests to avoid noise
	log := logger.NewZapLogger(zap.NewNop().Sugar())
	appService := services.NewInventoryApplicationService(repo, mockPublisher, log)
	eventHandlers := consumer.NewEventHandlers(appService, log)

	// Test Case 1: OrderCreated -> StockReserved
	t.Run("OrderCreated_ShouldReserveStock", func(t *testing.T) {
		// Clean database before test
		cleanDatabase(t, db)

		// Create test data for this test
		product := createTestProduct(t, ctx, appService)
		addTestStock(t, ctx, appService, product.ID, 100)

		// Arrange
		orderID := "order-123"
		orderCreatedEvent := &events.OrderCreated{
			OrderId: orderID,
			Items: []*common.OrderItem{
				{
					ProductId: product.ID,
					Quantity:  5,
				},
			},
		}

		// Configure mock publisher to expect PublishFromOutbox call
		mockPublisher.EXPECT().
			PublishFromOutbox(gomock.Any(), gomock.Any()).
			Return(nil).
			AnyTimes()

		// Act
		err := eventHandlers.HandleOrderCreated(ctx, orderCreatedEvent)

		// Assert
		require.NoError(t, err)

		// Check that stock is reserved
		stock, err := appService.GetStockByProductID(ctx, product.ID)
		require.NoError(t, err)
		assert.Equal(t, int32(95), stock.AvailableQuantity) // 100 - 5
		assert.Equal(t, int32(5), stock.ReservedQuantity)

		// Check that event is saved to outbox
		require.Eventually(t, func() bool {
			outboxEvents := getOutboxEvents(t, ctx, db)
			return len(outboxEvents) == 1 &&
				outboxEvents[0].Type == "StockReserved" &&
				outboxEvents[0].AggregateID == orderID
		}, 1*time.Second, 10*time.Millisecond, "StockReserved event should be saved to outbox")
	})

	// Test Case 2: PaymentProcessed (Success) -> StockCommitted
	t.Run("PaymentProcessed_Success_ShouldCommitStock", func(t *testing.T) {
		// Clean database before test
		cleanDatabase(t, db)

		// Create test data for this test
		product := createTestProduct(t, ctx, appService)
		addTestStock(t, ctx, appService, product.ID, 100)

		// Arrange
		orderID := "order-456"
		paymentProcessedEvent := &events.PaymentProcessed{
			OrderId: orderID,
			Success: true,
			Items: []*common.OrderItem{
				{
					ProductId: product.ID,
					Quantity:  3,
				},
			},
		}

		// Pre-reserve stock (this will create stock if it doesn't exist)
		reserveStockForOrder(t, ctx, appService, orderID, product.ID, 3)

		// Configure mock publisher to expect PublishFromOutbox call
		mockPublisher.EXPECT().
			PublishFromOutbox(gomock.Any(), gomock.Any()).
			Return(nil).
			AnyTimes()

		// Act
		err := eventHandlers.HandlePaymentProcessed(ctx, paymentProcessedEvent)

		// Assert
		require.NoError(t, err)

		// Check that reserved stock is committed
		stock, err := appService.GetStockByProductID(ctx, product.ID)
		require.NoError(t, err)
		assert.Equal(t, int32(97), stock.AvailableQuantity) // 100 - 3 (stock was created with 100, then 3 was reserved and committed)
		assert.Equal(t, int32(0), stock.ReservedQuantity)   // 3 - 3 (all reserved stock was committed)

		// Check that event is saved to outbox
		require.Eventually(t, func() bool {
			outboxEvents := getOutboxEvents(t, ctx, db)
			stockCommittedEvents := filterEventsByType(outboxEvents, "StockCommitted")
			return len(stockCommittedEvents) == 1 && stockCommittedEvents[0].AggregateID == orderID
		}, 1*time.Second, 10*time.Millisecond, "StockCommitted event should be saved to outbox")
	})

	// Test Case 3: PaymentProcessed (Failure) -> StockReleased
	t.Run("PaymentProcessed_Failure_ShouldReleaseStock", func(t *testing.T) {
		// Clean database before test
		cleanDatabase(t, db)

		// Create test data for this test
		product := createTestProduct(t, ctx, appService)
		addTestStock(t, ctx, appService, product.ID, 100)

		// Arrange
		orderID := "order-789"
		paymentProcessedEvent := &events.PaymentProcessed{
			OrderId: orderID,
			Success: false,
			Items: []*common.OrderItem{
				{
					ProductId: product.ID,
					Quantity:  2,
				},
			},
		}

		// Pre-reserve stock (this will create stock if it doesn't exist)
		reserveStockForOrder(t, ctx, appService, orderID, product.ID, 2)

		// Configure mock publisher to expect PublishFromOutbox call
		mockPublisher.EXPECT().
			PublishFromOutbox(gomock.Any(), gomock.Any()).
			Return(nil).
			AnyTimes()

		// Act
		err := eventHandlers.HandlePaymentProcessed(ctx, paymentProcessedEvent)

		// Assert
		require.NoError(t, err)

		// Check that reserved stock is released
		stock, err := appService.GetStockByProductID(ctx, product.ID)
		require.NoError(t, err)
		assert.Equal(t, int32(100), stock.AvailableQuantity) // 100 (stock was created with 100, then 2 was reserved and released back)
		assert.Equal(t, int32(0), stock.ReservedQuantity)    // 2 - 2 (all reserved stock was released)

		// Check that event is saved to outbox
		require.Eventually(t, func() bool {
			outboxEvents := getOutboxEvents(t, ctx, db)
			stockReleasedEvents := filterEventsByType(outboxEvents, "StockReleased")
			return len(stockReleasedEvents) == 1 && stockReleasedEvents[0].AggregateID == orderID
		}, 1*time.Second, 10*time.Millisecond, "StockReleased event should be saved to outbox")
	})

	// Test Case 4: Concurrent Order Processing
	t.Run("ConcurrentOrderProcessing_ShouldHandleRaceConditions", func(t *testing.T) {
		// Clean database before test
		cleanDatabase(t, db)

		// Create test data for this test
		product := createTestProduct(t, ctx, appService)
		addTestStock(t, ctx, appService, product.ID, 20) // Limited stock for concurrency test

		// Arrange
		orderID1 := "concurrent-order-1"
		orderID2 := "concurrent-order-2"
		orderCreatedEvent1 := &events.OrderCreated{
			OrderId: orderID1,
			Items: []*common.OrderItem{
				{
					ProductId: product.ID,
					Quantity:  10,
				},
			},
		}
		orderCreatedEvent2 := &events.OrderCreated{
			OrderId: orderID2,
			Items: []*common.OrderItem{
				{
					ProductId: product.ID,
					Quantity:  15,
				},
			},
		}

		// Configure mock publisher to expect PublishFromOutbox call
		mockPublisher.EXPECT().
			PublishFromOutbox(gomock.Any(), gomock.Any()).
			Return(nil).
			AnyTimes()

		// Act - run event processing in parallel
		done := make(chan error, 2)
		go func() {
			done <- eventHandlers.HandleOrderCreated(ctx, orderCreatedEvent1)
		}()
		go func() {
			done <- eventHandlers.HandleOrderCreated(ctx, orderCreatedEvent2)
		}()

		// Wait for both goroutines to complete
		err1 := <-done
		err2 := <-done

		// Assert
		// One order should succeed, the other should get insufficient stock error
		successCount := 0
		if err1 == nil {
			successCount++
		}
		if err2 == nil {
			successCount++
		}
		assert.Equal(t, 1, successCount, "Only one order should succeed due to insufficient stock")

		// Check final stock state
		require.Eventually(t, func() bool {
			stock, err := appService.GetStockByProductID(ctx, product.ID)
			if err != nil {
				return false
			}
			return stock.AvailableQuantity >= 0
		}, 1*time.Second, 10*time.Millisecond, "stock should not go negative")
	})
}

// TestOutboxPatternReliability tests outbox pattern reliability
func TestOutboxPatternReliability(t *testing.T) {
	// Setup
	ctx := context.Background()
	db, cleanup := setupTestDatabase(t, ctx)
	defer cleanup()

	// Create mock publisher that will fail
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	failingPublisher := mocks.NewMockEventPublisher(ctrl)

	// Create real components
	repo := repository.NewInventoryRepository(db)
	// Create a simple logger for tests
	// Use silent logger for tests to avoid noise
	log := logger.NewZapLogger(zap.NewNop().Sugar())
	appService := services.NewInventoryApplicationService(repo, failingPublisher, log)

	// Test Case 1: Event is saved to outbox when publisher fails
	t.Run("EventSavedToOutbox_WhenPublisherFails", func(t *testing.T) {
		// Clean database before test
		cleanDatabase(t, db)

		// Create test data for this test
		product := createTestProduct(t, ctx, appService)
		addTestStock(t, ctx, appService, product.ID, 50)

		// Arrange
		orderID := "outbox-test-order"
		req := &dto.ReserveStockRequest{
			OrderID: orderID,
			Items: []dto.StockReservationItem{
				{
					ProductID: product.ID,
					Quantity:  5,
				},
			},
		}

		// Configure mock publisher to return error
		failingPublisher.EXPECT().
			PublishFromOutbox(gomock.Any(), gomock.Any()).
			Return(errors.New("mock publisher failure")).
			AnyTimes()

		// Act
		_, err := appService.ReserveStock(ctx, req)

		// Assert
		require.NoError(t, err, "ReserveStock should succeed even if publisher fails")

		// Check that event is saved to outbox
		require.Eventually(t, func() bool {
			outboxEvents := getOutboxEvents(t, ctx, db)
			return len(outboxEvents) == 1 &&
				outboxEvents[0].Type == "StockReserved" &&
				outboxEvents[0].AggregateID == orderID &&
				outboxEvents[0].ProcessedAt == nil
		}, 1*time.Second, 10*time.Millisecond, "StockReserved event should be saved to outbox but not processed")
	})

	// Test Case 2: Background publisher processes events
	t.Run("BackgroundPublisher_ProcessesOutboxEvents", func(t *testing.T) {
		// Clean database before test
		cleanDatabase(t, db)

		// Create test data for this test
		product := createTestProduct(t, ctx, appService)
		addTestStock(t, ctx, appService, product.ID, 50)

		// Create an event in outbox first
		orderID := "background-test-order"
		req := &dto.ReserveStockRequest{
			OrderID: orderID,
			Items: []dto.StockReservationItem{
				{
					ProductID: product.ID,
					Quantity:  5,
				},
			},
		}

		// Reserve stock to create event in outbox
		_, err := appService.ReserveStock(ctx, req)
		require.NoError(t, err)

		// Arrange - create publisher that doesn't fail
		workingCtrl := gomock.NewController(t)
		defer workingCtrl.Finish()

		// Create mock kafka publisher
		mockKafkaPublisher := mocks.NewMockPublisher(workingCtrl)

		// Create background publisher
		backgroundPublisher := outbox.NewPublisher(
			repo,
			mockKafkaPublisher,
			"test-topic",
			log,
			10,
			100*time.Millisecond, // fast interval for tests
		)

		// Configure mock kafka publisher for successful publishing
		mockKafkaPublisher.EXPECT().
			Publish(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil).
			AnyTimes()

		// Act - start background publisher
		ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()

		go backgroundPublisher.Start(ctx)

		// Assert - wait for event processing
		require.Eventually(t, func() bool {
			outboxEvents := getOutboxEvents(t, ctx, db)
			return len(outboxEvents) == 1 && outboxEvents[0].ProcessedAt != nil
		}, 2*time.Second, 50*time.Millisecond, "event should be processed eventually")
	})

	// Test Case 3: Outbox retry mechanism - fails first, then recovers
	t.Run("OutboxRetryMechanism_ShouldRetryAfterFailure", func(t *testing.T) {
		// Clean database before test
		cleanDatabase(t, db)

		// Create test data for this test
		product := createTestProduct(t, ctx, appService)
		addTestStock(t, ctx, appService, product.ID, 50)

		// Arrange - create publisher that fails first, then works
		retryCtrl := gomock.NewController(t)
		defer retryCtrl.Finish()

		mockKafkaPublisher := mocks.NewMockPublisher(retryCtrl)

		// Configure mock kafka publisher: first call fails, second succeeds
		gomock.InOrder(
			mockKafkaPublisher.EXPECT().
				Publish(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(errors.New("temporary failure")),
			mockKafkaPublisher.EXPECT().
				Publish(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil),
		)

		// Create background publisher with fast interval
		backgroundPublisher := outbox.NewPublisher(
			repo,
			mockKafkaPublisher,
			"test-topic",
			log,
			10,
			50*time.Millisecond, // very fast interval for tests
		)

		// First create event in outbox (with failing publisher)
		orderID := "retry-test-order"
		req := &dto.ReserveStockRequest{
			OrderID: orderID,
			Items: []dto.StockReservationItem{
				{
					ProductID: product.ID,
					Quantity:  3,
				},
			},
		}

		// Act - reserve stock (event is saved to outbox)
		_, err := appService.ReserveStock(ctx, req)
		require.NoError(t, err)

		// Check that event in outbox is not processed
		require.Eventually(t, func() bool {
			outboxEvents := getOutboxEvents(t, ctx, db)
			return len(outboxEvents) == 1 && outboxEvents[0].ProcessedAt == nil
		}, 1*time.Second, 10*time.Millisecond, "Event should be saved to outbox but not processed initially")

		// Start background publisher
		ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
		defer cancel()

		go backgroundPublisher.Start(ctx)

		// Assert - wait for event processing after retry
		require.Eventually(t, func() bool {
			outboxEvents := getOutboxEvents(t, ctx, db)
			return len(outboxEvents) == 1 && outboxEvents[0].ProcessedAt != nil
		}, 2*time.Second, 50*time.Millisecond, "event should be processed after retry")
	})
}

// Helper functions

func setupTestDatabase(t *testing.T, ctx context.Context) (*gorm.DB, func()) {
	// Start PostgreSQL container
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15.4",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "inventory_test",
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	// Get connection string
	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	dsn := "host=" + host + " port=" + port.Port() + " user=test password=test dbname=inventory_test sslmode=disable"

	// Connect to database with retry
	var db *gorm.DB
	for i := 0; i < 10; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		if i < 9 {
			time.Sleep(500 * time.Millisecond)
		}
	}
	require.NoError(t, err, "Failed to connect to database after retries")

	// Run migrations
	migrationService := migration.NewMigrationService(db)
	err = migrationService.Migrate(ctx)
	require.NoError(t, err)

	// Disable GORM logging in tests to reduce noise
	db.Logger = gormLogger.Default.LogMode(gormLogger.Silent)

	// Cleanup function
	cleanup := func() {
		if container != nil {
			container.Terminate(ctx)
		}
	}

	return db, cleanup
}

// cleanDatabase clears all tables in test database
func cleanDatabase(t *testing.T, db *gorm.DB) {
	// Clear tables using TRUNCATE for better performance and reliability
	tables := []string{"outbox_records", "stocks", "products"}

	for _, table := range tables {
		err := db.Exec("TRUNCATE TABLE " + table + " RESTART IDENTITY CASCADE").Error
		require.NoError(t, err)
	}
}

func createTestProduct(t *testing.T, ctx context.Context, appService *services.InventoryApplicationService) *dto.ProductResponse {
	currency, _ := valueobjects.NewCurrency("USD")
	req := &dto.CreateProductRequest{
		Name:     "Test Product",
		Price:    valueobjects.Money{Amount: 1000, Currency: currency},
		ImageURL: "https://example.com/image.jpg",
		Stock:    0, // Add stock separately
	}

	product, err := appService.CreateProduct(ctx, req)
	require.NoError(t, err)
	return product
}

func addTestStock(t *testing.T, ctx context.Context, appService *services.InventoryApplicationService, productID string, quantity int32) {
	_, err := appService.AddStock(ctx, productID, quantity)
	require.NoError(t, err)
}

func reserveStockForOrder(t *testing.T, ctx context.Context, appService *services.InventoryApplicationService, orderID, productID string, quantity int32) {
	req := &dto.ReserveStockRequest{
		OrderID: orderID,
		Items: []dto.StockReservationItem{
			{
				ProductID: productID,
				Quantity:  quantity,
			},
		},
	}

	_, err := appService.ReserveStock(ctx, req)
	require.NoError(t, err)
}

func getOutboxEvents(t *testing.T, ctx context.Context, db *gorm.DB) []outbox.Event {
	var records []struct {
		ID          uint
		AggregateID string
		Type        string
		Payload     string
		RetryCount  int
		Processed   bool
		ProcessedAt *time.Time
		FailedAt    *time.Time
		Error       string
		CreatedAt   time.Time
		UpdatedAt   time.Time
	}

	err := db.Table("outbox_records").Find(&records).Error
	require.NoError(t, err)

	events := make([]outbox.Event, len(records))
	for i, record := range records {
		events[i] = outbox.Event{
			ID:          record.ID,
			AggregateID: record.AggregateID,
			Type:        record.Type,
			Payload:     record.Payload,
			RetryCount:  record.RetryCount,
			ProcessedAt: record.ProcessedAt,
			FailedAt:    record.FailedAt,
			Error:       record.Error,
			CreatedAt:   record.CreatedAt,
		}
	}

	return events
}

func filterEventsByType(events []outbox.Event, eventType string) []outbox.Event {
	var filtered []outbox.Event
	for _, event := range events {
		if event.Type == eventType {
			filtered = append(filtered, event)
		}
	}
	return filtered
}
