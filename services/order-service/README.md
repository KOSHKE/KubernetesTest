# Order Service

## Overview

The Order Service is a core microservice within the e-commerce platform responsible for order management, processing, and lifecycle management. It handles the complete order workflow from creation to completion, including order validation, item management, status tracking, and integration with payment and inventory services.

**Role in the Platform:**
- Centralized order management and processing
- Order lifecycle orchestration (creation, modification, cancellation)
- Integration point for payment and inventory services
- Order status tracking and business rule enforcement

## Architecture

### Layered Architecture

The service follows Domain-Driven Design (DDD) principles with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                Infrastructure Layer                         │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐          │
│  │   gRPC      │ │ Repository  │ │   Kafka     │          │
│  │   Server    │ │   (GORM)    │ │ Publisher   │          │
│  │ Consumers   │ │ Migration   │ │ Consumers   │          │
│  └─────────────┘ └─────────────┘ └─────────────┘          │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                Application Layer                            │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐          │
│  │   Use       │ │ Application │ │     DTOs    │          │
│  │   Cases     │ │  Services   │ │             │          │
│  └─────────────┘ └─────────────┘ └─────────────┘          │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                  Domain Layer                               │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐          │
│  │ Aggregates  │ │   Ports     │ │ Value      │          │
│  │ (Order)     │ │(Interfaces) │ │ Objects    │          │
│  │ Entities    │ │ Services    │ │            │          │
│  └─────────────┘ └─────────────┘ └─────────────┘          │
└─────────────────────────────────────────────────────────────┘
```

### Technology Stack

- **Language**: Go 1.21+
- **Communication**: gRPC with Protocol Buffers
- **Database**: PostgreSQL with GORM ORM
- **Event Streaming**: Apache Kafka for event publishing and consuming
- **Metrics**: Prometheus-compatible metrics with HTTP endpoint
- **Logging**: Structured logging with Zap
- **Configuration**: Environment-based configuration management

### Dependency Organization

```
Order Service
├── Domain Layer
│   ├── Aggregates (Order aggregate with business rules)
│   ├── Entities (OrderItem entity)
│   ├── Value Objects (OrderItem, OrderStatus, ShippingAddress)
│   ├── Ports (Repository, Publisher, Consumer interfaces)
│   └── Services (OrderDomainService for business logic)
├── Application Layer
│   ├── Use Cases (Create, Get, Update, Cancel, AddItem, RemoveItem)
│   ├── Services (OrderApplicationService, PaymentEventService, StockEventService)
│   └── DTOs (Request/Response models)
├── Infrastructure Layer
│   ├── Repository (GORM implementation)
│   ├── Kafka Publisher (event publishing)
│   ├── Kafka Consumers (event processing)
│   ├── gRPC Server
│   └── Migration Service
└── Cross-cutting Concerns
    ├── Metrics (Prometheus)
    ├── Logging (Zap)
    └── Error Handling (domain-specific errors)
```

## Data Flow

### Request Flow

```
Client Request → gRPC Server → Application Service → Use Case → Repository → Database
                                                                           ↓
Response ← gRPC Server ← Application Service ← Use Case ← Repository ← Database
```

### Event Publishing Flow

```
Order Created → Kafka Publisher → Payment Service (OrderCreated event)
Payment Processed → Order Service (PaymentProcessed event)
Stock Reserved → Order Service (StockReserved event)
```

### Order Lifecycle

```
1. Order Creation → Validate Items → Create Order → Publish OrderCreated
2. Payment Processing → Update Status to PAID/PAYMENT_FAILED
3. Stock Reservation → Update Status to CONFIRMED
4. Order Status Management → PENDING → CONFIRMED → CANCELLED
```

### Event Processing Flow

```
PaymentProcessed Event → PaymentEventService → OrderDomainService → Update Order Status
StockReserved Event → StockEventService → OrderDomainService → Update Order Status
```

## Components

### Order Aggregate

**Location**: `internal/domain/aggregates/order.go`

The Order aggregate encapsulates all business rules and operations related to orders:

**Business Rules:**
- Each product can appear only once per order
- If the same product is added again, quantity is updated
- All items in order must have the same currency as the order
- Only pending orders can be cancelled
- Users can only cancel their own orders

**Key Methods:**
- `NewOrder(userID, shippingAddress, currency)` - Create new order
- `AddItem(productID, productName, quantity, unitPrice)` - Add/update item
- `RemoveItem(productID)` - Remove item from order
- `CancelOrder(userID)` - Cancel order with access control
- `SetStatus(newStatus)` - Update order status
- `IsOwnedBy(userID)` - Check ownership

### Order Entity

**Location**: `internal/domain/entities/order_item.go`

The OrderItem entity represents individual items within an order:

**Key Methods:**
- `NewOrderItem(productID, productName, quantity, unitPrice)` - Create new item
- `ChangeQuantity(newQuantity)` - Update item quantity
- `TotalPrice()` - Calculate total price for the item
- `Currency()` - Get item currency
- `UnitPriceAmount()` - Get unit price in minor units
- `TotalPriceAmount()` - Get total price in minor units

### Order Repository (GORM)

**Location**: `internal/infra/repository/gorm_order_repository.go`

- Implements `OrderRepository` interface
- Uses GORM for database operations
- Separates domain aggregates from database records
- Supports transactions via `WithTx()` method
- Handles CRUD operations for order management

**Key Methods:**
- `Create(ctx, order)` - Create new order
- `GetByID(ctx, id)` - Retrieve order by ID
- `GetByUserID(ctx, userID, page, limit)` - Get user's orders with pagination
- `Update(ctx, order)` - Update order information

### Value Objects

**Location**: `internal/domain/valueobjects/`

1. **OrderItem** - Represents an item in an order as a value object
   - ProductID, ProductName, Quantity, Price
   - Immutable value object with validation
   - `Equals()` method for comparison

2. **OrderStatus** - Order status enumeration
   - PENDING, CONFIRMED, CANCELLED, PAID, PAYMENT_FAILED
   - Type-safe status management
   - String representation support

3. **ShippingAddress** - Delivery address
   - String-based value object with validation
   - Immutable and validated
   - `IsEmpty()` method for validation

### Use Cases

**Location**: `internal/application/usecases/`

1. **CreateOrderUseCase**
   - Validates order data and items
   - Creates order aggregate via factory
   - Publishes OrderCreated event
   - Records metrics for order creation

2. **GetOrderUseCase**
   - Retrieves order by ID
   - Validates user access to order
   - Returns order details

3. **GetUserOrdersUseCase**
   - Retrieves paginated list of user's orders
   - Supports filtering and sorting
   - Returns order summaries

4. **UpdateOrderStatusUseCase**
   - Updates order status with validation
   - Enforces business rules for status transitions
   - Publishes status change events

5. **CancelOrderUseCase**
   - Cancels order with access control
   - Uses domain method for business rule validation
   - Publishes cancellation events

6. **AddItemToOrderUseCase**
   - Adds new item to existing order
   - Updates quantities for existing products
   - Recalculates order total

7. **RemoveItemFromOrderUseCase**
   - Removes item from order
   - Recalculates order total
   - Validates order state

8. **ProcessOrderUseCase**
   - Handles order processing workflow
   - Coordinates with external services
   - Manages order state transitions

### Application Services

**Location**: `internal/application/services/`

1. **OrderApplicationService**
   - Orchestrates order operations
   - Provides unified interface for gRPC server
   - Handles DTO conversion and validation
   - Coordinates use cases

2. **PaymentEventService**
   - Processes PaymentProcessed events
   - Updates order status based on payment results
   - Handles payment success and failure scenarios

3. **StockEventService**
   - Processes StockReserved events
   - Confirms stock reservation
   - Updates order status to CONFIRMED

### Kafka Publisher

**Location**: `internal/infra/publisher/`

- Publishes domain events to Kafka topics
- Ensures reliable event delivery
- Handles event serialization and routing
- Critical for service integration

**Events Published:**
- `OrderCreated` - New order created

### Kafka Consumers

**Location**: `internal/infra/consumer/`

1. **PaymentProcessedConsumer**
   - Consumes PaymentProcessed events
   - Updates order status based on payment results
   - Handles payment success and failure scenarios

2. **StockReservedConsumer**
   - Consumes StockReserved events
   - Confirms stock reservation
   - Updates order status to CONFIRMED

### gRPC Server

**Location**: `internal/infra/grpc/`

- Protocol Buffer-based gRPC server
- Health check service integration
- Error mapping to gRPC status codes
- Request/response DTO mapping

**Services:**
- `CreateOrder` - Create new order
- `GetOrder` - Retrieve order details
- `GetUserOrders` - Get user's order list
- `UpdateOrderStatus` - Update order status
- `CancelOrder` - Cancel order
- `AddItemToOrder` - Add item to order
- `RemoveItemFromOrder` - Remove item from order

### Domain Services

**Location**: `internal/domain/services/`

1. **OrderDomainService**
   - Handles order business logic in domain layer
   - Confirms payment and updates status to PAID
   - Marks payment as failed
   - Confirms stock reservation and updates status to CONFIRMED

### Migration Service

**Location**: `internal/infra/migration/`

- Database schema initialization
- Automatic migration execution on startup
- Ensures database consistency
- Order and order_items table management

### Metrics

**Location**: `internal/metrics/`

- Prometheus-compatible metrics endpoint
- Order-specific business metrics
- HTTP metrics (request count, duration, errors)
- Separate port for metrics collection

**Business Metrics:**
- `order_created_total` - Total orders created
- `order_creation_failed_total` - Total order creation failures
- Entity events with order entity type and actions

## Lifecycle

### Application Initialization

```
main() → Run() → initialize() → start() → waitForShutdown()
```

**Initialize Phase:**
1. **Infrastructure**: Database connection, migrations, Kafka publisher, metrics server
2. **Business Logic**: Repository, use cases, application service
3. **Kafka Components**: Publishers and consumers initialization
4. **gRPC Server**: Server setup, service registration, health checks

**Start Phase:**
- Start gRPC server on configured port
- Start metrics HTTP server in background
- Start Kafka consumers in background
- Application ready to serve requests

### Graceful Shutdown

**Shutdown Sequence:**
1. **Signal Handling**: OS interrupt signals (SIGTERM, SIGINT, SIGHUP)
2. **gRPC Server**: Graceful stop with 5-second timeout
3. **Resource Cleanup**: Close database connections, Kafka connections
4. **Context Cancellation**: Cancel all background operations

**Shutdown Timeout:**
- gRPC graceful shutdown: 5 seconds
- Force stop if graceful shutdown fails

## Business Rules & Domain Logic

### Order Creation Rules

- User must be authenticated
- Order must contain at least one item
- All items must have valid product information
- Shipping address must be provided
- Currency must be consistent across all items

### Order Modification Rules

- Only pending orders can be modified
- Users can only modify their own orders
- Item quantities must be positive
- Currency consistency must be maintained

### Order Cancellation Rules

- Only pending orders can be cancelled
- Users can only cancel their own orders
- Cancellation is irreversible
- Inventory must be released upon cancellation

### Status Transition Rules

```
PENDING → PAID (payment processed successfully)
PENDING → PAYMENT_FAILED (payment failed)
PENDING → CANCELLED (user cancellation)
PAID → CONFIRMED (stock reserved)
```

## Extensibility & Design Patterns

### Domain-Driven Design (DDD)

- **Domain Layer**: Core business logic and aggregates
- **Application Layer**: Use cases and orchestration
- **Infrastructure Layer**: External concerns (database, gRPC, Kafka)

### SOLID Principles

- **Single Responsibility**: Each use case handles one business operation
- **Open/Closed**: Extensible through interface implementations
- **Liskov Substitution**: Repository and publisher interfaces
- **Interface Segregation**: Focused interfaces for different concerns
- **Dependency Inversion**: High-level modules depend on abstractions

### Clean Architecture

- **Use Cases**: Operate on domain objects, not DTOs
- **Application Services**: Handle DTO conversion and validation
- **Domain Layer**: Pure business logic without external dependencies
- **Infrastructure**: Implements domain interfaces

### Extension Points

**New Use Cases:**
1. Implement use case interface
2. Add to application service
3. Register in gRPC server
4. Add corresponding metrics

**New Order Statuses:**
1. Add to OrderStatus value object
2. Update status transition rules in aggregate
3. Add corresponding business logic
4. Update event publishing

**Additional Repositories:**
1. Implement repository interface
2. Register in dependency injection
3. Use in appropriate use cases

**New Events:**
1. Define event in proto files
2. Implement publisher interface
3. Add event publishing to use cases
4. Update event handling in other services

## Configuration Management

**Environment Variables:**
- `ORDER_SERVICE_PORT`: gRPC server port (default: 50052)
- `ORDER_SERVICE_METRICS_PORT`: Metrics HTTP port (default: 9095)
- `DATABASE_URL`: PostgreSQL connection URL
- `KAFKA_BROKERS`: Kafka broker addresses
- `ORDER_DEFAULT_CURRENCY`: Default currency (default: USD)
- `ORDER_MAX_ITEMS`: Maximum items per order (default: 100)
- `ORDER_MAX_VALUE`: Maximum order value in minor units (default: 1000000)
- `INVENTORY_SERVICE_URL`: Inventory service URL (default: inventory-service:50053)
- `INVENTORY_PROVIDER_TIMEOUT`: Inventory service timeout (default: 3s)
- `ORDER_ITEM_MAX_QUANTITY`: Maximum item quantity (default: 1000)
- `ORDER_ITEM_MIN_QUANTITY`: Minimum item quantity (default: 1)
- `ORDER_ITEM_MAX_PRICE`: Maximum item price in minor units (default: 100000)
- `ORDER_SUPPORTED_CURRENCIES`: Supported currencies (default: USD,EUR,GBP,RUB)
- `ORDER_GRACEFUL_SHUTDOWN_TIMEOUT`: Graceful shutdown timeout (default: 30s)

**Validation:**
- Required database and Kafka configuration
- Valid port numbers and URLs
- Proper timeout values
- Currency support validation

## Development

### Prerequisites

- Go 1.21+
- PostgreSQL
- Apache Kafka
- Protocol Buffers compiler

### Local Development

```bash
# Set environment variables
export DATABASE_URL="postgres://user:pass@localhost:5432/orders"
export KAFKA_BROKERS="localhost:9092"
export ORDER_SERVICE_PORT="50052"
export ORDER_SERVICE_METRICS_PORT="9095"

# Run service
go run cmd/main.go
```

### Docker

```bash
# Build and run with docker-compose
docker-compose up order-service
```

### Testing

```bash
# Run tests
go test ./...

# Run with coverage
go test -cover ./...
```

## Monitoring & Observability

### Metrics

**Business Metrics:**
- `order_created_total`: Total orders created
- `order_creation_failed_total`: Total order creation failures
- Entity events with order entity type and actions

**HTTP Metrics:**
- Request count and duration
- Error rates and status codes
- Response time percentiles

### Health Checks

- gRPC health check service
- Database connectivity verification
- Kafka connectivity verification

### Logging

- Structured JSON logging
- Request correlation IDs
- Error context and stack traces
- Performance metrics integration
- Event publishing logs

## Integration Points

### Payment Service Integration

- Publishes `OrderCreated` event for payment processing
- Receives `PaymentProcessed` event for order status updates
- Handles payment success and failure scenarios
- Updates order status to PAID or PAYMENT_FAILED

### Inventory Service Integration

- Publishes `OrderCreated` event for stock reservation
- Receives `StockReserved` event for order confirmation
- Updates order status to CONFIRMED after stock reservation
- Handles stock reservation failures

### User Service Integration

- Validates user authentication for order operations
- Enforces user ownership of orders
- Integrates with user profile data

## Error Handling

### Domain Errors

- `ErrOrderNotFound`: Order not found
- `ErrOrderAccessDenied`: User cannot access order
- `ErrOrderCancellationFailed`: Order cannot be cancelled
- `ErrOrderValidationFailed`: Order data validation failed
- `ErrOrderEventPublishFailed`: Event publishing failed
- `ErrOrderCreationFailed`: Order creation failed
- `ErrOrderRetrievalFailed`: Order retrieval failed
- `ErrOrderPersistenceFailed`: Order persistence failed
- `ErrOrderItemProcessingFailed`: Order item processing failed
- `ErrOrderItemNotFound`: Order item not found
- `ErrOrderItemAdditionFailed`: Order item addition failed
- `ErrOrderStatusUpdateFailed`: Order status update failed
- `ErrCurrencyMismatch`: Currency mismatch error

### Error Recovery

- Automatic retry for transient failures
- Graceful degradation for non-critical operations
- Event publishing failure handling
- Database transaction rollback

## Performance Considerations

### Database Optimization

- Indexed queries for user orders
- Efficient pagination for large order lists
- Optimized joins for order items

### Event Publishing

- Asynchronous event publishing
- Retry mechanisms for failed events
- Batch processing for high-volume scenarios

### Caching Strategy

- Order data caching for frequently accessed orders
- User order list caching
- Status-based filtering optimization
