# Inventory Service

## Overview

The Inventory Service is a core microservice within the e-commerce platform responsible for product catalog management, stock tracking, and inventory operations. It handles product information, stock levels, reservations, and integrates with order and payment services to manage inventory lifecycle through event-driven architecture.

**Role in the Platform:**
- Product catalog management and maintenance
- Stock level tracking and monitoring
- Stock reservation and release operations
- Event-driven integration with order and payment services
- Inventory business rule enforcement
- Reliable event publishing using outbox pattern

## Architecture

### Layered Architecture

The service follows Domain-Driven Design (DDD) principles with clear separation of concerns and event-driven architecture:

```
┌─────────────────────────────────────────────────────────────┐
│                Infrastructure Layer                         │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐          │
│  │   gRPC      │ │ Repository  │ │   Kafka     │          │
│  │   Server    │ │   Facade    │ │ Publisher   │          │
│  │ Consumers   │ │ (GORM)      │ │ Consumers   │          │
│  │ Migration   │ │ Outbox      │ │ Outbox      │          │
│  └─────────────┘ └─────────────┘ └─────────────┘          │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                Application Layer                            │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐          │
│  │   Use       │ │ Application │ │     DTOs    │          │
│  │   Cases     │ │  Services   │ │ (Typed)     │          │
│  └─────────────┘ └─────────────┘ └─────────────┘          │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                  Domain Layer                               │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐          │
│  │ Aggregates  │ │   Ports     │ │ Entities    │          │
│  │(ProductInv) │ │(Interfaces) │ │ (Product,   │          │
│  │             │ │             │ │  Stock)     │          │
│  └─────────────┘ └─────────────┘ └─────────────┘          │
└─────────────────────────────────────────────────────────────┘
```

### Technology Stack

- **Language**: Go 1.21+
- **Communication**: gRPC with Protocol Buffers
- **Database**: PostgreSQL with GORM ORM
- **Event Streaming**: Apache Kafka for event publishing and consuming
- **Outbox Pattern**: Reliable event publishing with background processing
- **Metrics**: Prometheus-compatible metrics with HTTP endpoint
- **Logging**: Structured logging with Zap
- **Configuration**: Environment-based configuration management

## Components

### ProductInventory Aggregate

**Location**: `internal/domain/aggregates/product_inventory.go`

The ProductInventory aggregate combines Product and Stock entities to solve the N+1 problem and provide a unified interface for inventory operations:

**Business Rules:**
- Products must have valid pricing information
- Stock quantities cannot be negative
- Stock reservations must be validated before committing
- Stock operations are atomic within transactions
- Aggregate validation delegates to entity validations

**Key Methods:**
- `Validate()` - Validates both product and stock entities
- `ToProduct()` - Converts aggregate to Product entity
- `ToStock()` - Converts aggregate to Stock entity

### Domain Entities

**Location**: `internal/domain/entities/`

1. **Product** - Represents a product entity
   - ID, Name, Price (Money value object)
   - Image URL and timestamps
   - Validation for required fields

2. **Stock** - Represents stock information as an entity
   - ProductID, AvailableQuantity, ReservedQuantity
   - Business logic for stock operations
   - Validation for quantity constraints

**Key Stock Methods:**
- `CanReserve(quantity)` - Check if quantity can be reserved
- `Reserve(quantity)` - Reserve stock for orders
- `Release(quantity)` - Release reserved stock back to available
- `Commit(quantity)` - Commit reserved stock (remove permanently)
- `AddQuantity(quantity)` - Add stock to available quantity
- `GetTotalQuantity()` - Get total stock (available + reserved)

### Use Cases

**Location**: `internal/application/usecases/`

1. **CreateProductUseCase**
   - Validates product data
   - Creates product entity
   - Handles product creation business rules

2. **GetProductUseCase**
   - Retrieves product by ID
   - Returns product entity

3. **AddStockUseCase**
   - Adds stock to existing products
   - Handles stock creation and updates

4. **ReserveStockUseCase**
   - Reserves stock for orders using transactions
   - Uses batch operations for performance
   - Saves events to outbox for reliable publishing
   - Handles race conditions with FOR UPDATE locks

5. **ReleaseStockUseCase**
   - Releases reserved stock back to available
   - Publishes release events via outbox

6. **CommitStockUseCase**
   - Commits reserved stock (removes permanently)
   - Publishes commit events via outbox

### Application Services

**Location**: `internal/application/services/`

1. **InventoryApplicationService**
   - Orchestrates inventory operations
   - Provides unified interface for gRPC server
   - Coordinates use cases and repositories
   - Handles DTO conversions
   - Implements batch operations for performance
   - Uses ProductInventory aggregate to solve N+1 problem

### Data Transfer Objects (DTOs)

**Location**: `internal/application/dto/`

1. **Request DTOs**
   - `CreateProductRequest` - Product creation with stock
   - `ReserveStockRequest` - Stock reservation for orders
   - `ReleaseStockRequest` - Stock release for cancellations
   - `CommitStockRequest` - Stock commit for payments
   - `ListProductsRequest` - Paginated product listing

2. **Response DTOs**
   - `ProductResponse` - Product with stock information
   - `StockInfo` - Stock quantity details
   - `ListProductsResponse` - Paginated product list
   - `ReserveStockResponse` - Reservation result
   - `ReleaseStockResponse` - Release result
   - `CommitStockResponse` - Commit result

3. **Event DTOs**
   - `StockEventDTO` - Stock events for outbox pattern

### Infrastructure Components

**Location**: `internal/infra/`

1. **Repository Facade (GORM)**
   - Implements `InventoryRepositoryFacade` interface
   - Combines ProductRepository, StockRepository, and OutboxRepository
   - Supports transactions across all repositories
   - Uses factory pattern for repository creation
   - Implements batch operations for performance

2. **gRPC Server**
   - Protocol Buffer-based gRPC server
   - Health check service integration
   - Error mapping to gRPC status codes
   - Reflection support for development

3. **Kafka Publisher**
   - Publishes stock-related events
   - Handles event serialization and routing
   - Typed event publishing

4. **Kafka Consumers**
   - Consumes order and payment events
   - Event-driven stock management
   - Handles OrderCreated and PaymentProcessed events
   - Consumer manager for lifecycle management

5. **Outbox Pattern**
   - Reliable event publishing
   - Background publisher for event processing
   - Ensures at-least-once delivery

6. **Migration Service**
   - Database schema initialization
   - Auto-migration for development
   - Product, Stock, and Outbox table management

## Business Rules & Domain Logic

### Stock Management Rules

- Stock quantities cannot be negative
- Only available stock can be reserved
- Reserved stock can be released or committed
- Stock operations are atomic within transactions
- Batch operations for performance optimization
- FOR UPDATE locks prevent race conditions

### Product Management Rules

- Products must have valid pricing (Money value object)
- Product names are required
- Product IDs are generated using ID generator
- Product information can be updated while maintaining referential integrity

### Event Publishing Rules

- Stock reserved events are published via outbox pattern
- Stock released events are published on order cancellation
- Stock committed events are published on successful payment
- Events are published asynchronously with retry mechanisms
- At-least-once delivery guarantee

### Transaction Management

- All stock operations use database transactions
- Outbox events are saved within the same transaction
- Repository facade ensures transaction consistency
- Rollback on any operation failure

## Configuration

**Environment Variables:**
- `INVENTORY_SERVICE_PORT`: gRPC server port (default: 50053)
- `INVENTORY_SERVICE_METRICS_PORT`: Metrics HTTP port (default: 9096)
- `DATABASE_URL`: PostgreSQL connection URL
- `KAFKA_BROKERS`: Kafka broker addresses (comma-separated)
- `INVENTORY_DEFAULT_CURRENCY`: Default currency (default: USD)
- `INVENTORY_MAX_PRODUCTS_PER_PAGE`: Max products per page (default: 100)
- `INVENTORY_SUPPORTED_CURRENCIES`: Supported currencies (default: USD,EUR,GBP,RUB)

**Configuration Structure:**
- Uses centralized configuration from `pkg/config`
- Environment-based configuration loading
- Validation of required configuration values
- Default values for optional settings

## Integration Points

### Order Service Integration

- Consumes `OrderCreated` events for stock reservation
- Publishes `StockReserved` events via outbox pattern
- Handles stock release on order cancellation
- Event-driven architecture with reliable processing

### Payment Service Integration

- Consumes `PaymentProcessed` events
- Commits reserved stock on successful payment
- Releases reserved stock on payment failure
- Event handlers process payment results asynchronously

### Event Flow

1. **Order Creation**: Order service publishes `OrderCreated` → Inventory reserves stock → Publishes `StockReserved`
2. **Payment Success**: Payment service publishes `PaymentProcessed` (success) → Inventory commits stock → Publishes `StockCommitted`
3. **Payment Failure**: Payment service publishes `PaymentProcessed` (failure) → Inventory releases stock → Publishes `StockReleased`

## Error Handling

### Domain Errors

- `ErrProductNotFound`: Product not found
- `ErrInsufficientStock`: Not enough stock available
- `ErrInvalidProductID`: Invalid product identifier
- `ErrInvalidQuantity`: Invalid quantity value
- `ErrInvalidProductName`: Invalid product name
- `ErrInsufficientReservedStock`: Not enough reserved stock

### Error Handling Strategy

- Domain-specific errors for business logic failures
- gRPC status code mapping for API responses
- Structured error logging with context
- Graceful error handling in event consumers
- Transaction rollback on errors

## Performance Considerations

### Database Optimization

- Indexed queries for product and stock lookups
- Efficient pagination for product listings
- Batch operations for stock management
- FOR UPDATE locks for race condition prevention
- ProductInventory aggregate solves N+1 problem
- Single query for products with stock information

### Event Processing

- Asynchronous event publishing via outbox pattern
- Background publisher with configurable intervals
- Retry mechanisms for failed events
- Batch processing for high-volume scenarios
- Consumer manager for efficient resource usage

### Transaction Management

- Database transactions for data consistency
- Repository facade pattern for transaction coordination
- Optimistic locking for concurrent operations
- Minimal transaction scope for better performance

## Development

### Prerequisites

- Go 1.21+
- PostgreSQL
- Apache Kafka
- Protocol Buffers compiler
- Docker (for containerized development)

### Local Development

```bash
# Set environment variables
export DATABASE_URL="postgres://user:pass@localhost:5432/inventory"
export KAFKA_BROKERS="localhost:9092"
export INVENTORY_SERVICE_PORT="50053"
export INVENTORY_SERVICE_METRICS_PORT="9096"

# Run service
go run cmd/main.go
```

### Docker Development

```bash
# Build and run with docker-compose
docker-compose up inventory-service

# Or run specific service
docker-compose up -d postgres kafka
docker-compose up inventory-service
```

### Testing

```bash
# Run unit tests
go test ./...

# Run integration tests
go test ./tests/integration/...

# Run with coverage
go test -cover ./...
```

### Metrics and Monitoring

- **Metrics Endpoint**: `http://localhost:9096/metrics`
- **Health Check**: `http://localhost:9096/healthz`
- **Readiness Check**: `http://localhost:9096/readyz`
- **Profiling**: `http://localhost:6060/debug/pprof/`

## Metrics

### Business Metrics

The service provides comprehensive metrics for monitoring inventory operations:

- **Product Metrics**
  - `product_created_total` - Total products created
  - `product_creation_failed_total` - Product creation failures

- **Stock Metrics**
  - `stock_reserved_total` - Stock reservations
  - `stock_reservation_failed_total` - Reservation failures
  - `stock_released_total` - Stock releases
  - `stock_committed_total` - Stock commits

- **HTTP Metrics**
  - Request duration, status codes, and throughput
  - gRPC method call metrics
  - Error rates and response times

### Metrics Implementation

- Uses Prometheus-compatible metrics
- Structured metric labels for filtering
- Integration with centralized metrics server
- Business-specific metrics for inventory operations
