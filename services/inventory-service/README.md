# Inventory Service

## Overview

The Inventory Service is a core microservice within the e-commerce platform responsible for product catalog management, stock tracking, and inventory operations. It handles product information, stock levels, reservations, and integrates with order and payment services to manage inventory lifecycle.

**Role in the Platform:**
- Product catalog management and maintenance
- Stock level tracking and monitoring
- Stock reservation and release operations
- Integration point for order and payment services
- Inventory business rule enforcement

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
│  │ (Product)   │ │(Interfaces) │ │ Objects    │          │
│  │ Entities    │ │ Services    │ │ (Stock)    │          │
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

## Components

### Product Aggregate

**Location**: `internal/domain/aggregates/product.go`

The Product aggregate encapsulates all business rules and operations related to products and stock:

**Business Rules:**
- Products must have valid pricing information
- Stock quantities cannot be negative
- Stock reservations must be validated before committing
- Only active products can be reserved
- Stock operations are atomic within the aggregate

**Key Methods:**
- `ReserveStock(quantity)` - Reserve stock for an order
- `ReleaseStock(quantity)` - Release reserved stock
- `CommitStock(quantity)` - Commit reserved stock (remove permanently)
- `AddStock(quantity)` - Add stock to available quantity
- `IsAvailableForPurchase()` - Check if product can be purchased

### Domain Entities

**Location**: `internal/domain/entities/`

1. **Product** - Represents a product entity
   - ID, Name, Description, Price
   - Image URL and active status
   - Creation and update timestamps

### Value Objects

**Location**: `internal/domain/valueobjects/`

1. **Stock** - Represents stock information as a value object
   - Available and reserved quantities
   - Immutable operations for stock management
   - Business logic for stock validation

### Use Cases

**Location**: `internal/application/usecases/`

1. **CreateProductUseCase**
   - Validates product data
   - Creates product aggregate
   - Handles product creation business rules

2. **GetProductUseCase**
   - Retrieves product by ID
   - Returns product with current stock information

3. **ReserveStockUseCase**
   - Reserves stock for orders
   - Publishes stock events
   - Handles reservation failures

### Application Services

**Location**: `internal/application/services/`

1. **InventoryApplicationService**
   - Orchestrates inventory operations
   - Provides unified interface for gRPC server
   - Coordinates use cases and domain services

### Domain Services

**Location**: `internal/domain/services/`

1. **InventoryDomainService**
   - Handles complex inventory business logic
   - Manages stock reservations and releases
   - Coordinates with external services

### Infrastructure Components

**Location**: `internal/infra/`

1. **Repository (GORM)**
   - Implements `InventoryRepository` interface
   - Handles database operations for products, categories, and stock
   - Supports transactions

2. **gRPC Server**
   - Protocol Buffer-based gRPC server
   - Health check service integration
   - Error mapping to gRPC status codes

3. **Kafka Publisher**
   - Publishes stock-related events
   - Handles event serialization and routing

4. **Kafka Consumers**
   - Consumes order and payment events
   - Updates stock based on order lifecycle

5. **Migration Service**
   - Database schema initialization
   - Seed data management

## Business Rules & Domain Logic

### Stock Management Rules

- Stock reservations are temporary and have TTL
- Only available stock can be reserved
- Reserved stock can be released or committed
- Stock operations are atomic within transactions

### Product Management Rules

- Products must have valid pricing
- Categories must be active to be assigned to products
- Product information can be updated while maintaining referential integrity

### Event Publishing Rules

- Stock reserved events are published on successful reservation
- Stock reservation failed events are published on failures
- Stock released events are published on order cancellation
- Stock committed events are published on successful payment

## Configuration

**Environment Variables:**
- `INVENTORY_SERVICE_PORT`: gRPC server port (default: 50053)
- `INVENTORY_SERVICE_METRICS_PORT`: Metrics HTTP port (default: 9096)
- `DATABASE_URL`: PostgreSQL connection URL
- `KAFKA_BROKERS`: Kafka broker addresses
- `INVENTORY_DEFAULT_CURRENCY`: Default currency (default: USD)
- `INVENTORY_MAX_PRODUCTS_PER_PAGE`: Max products per page (default: 100)
- `INVENTORY_RESERVATION_TTL_SECONDS`: Reservation TTL (default: 900)
- `INVENTORY_SUPPORTED_CURRENCIES`: Supported currencies (default: USD,EUR,GBP,RUB)

## Integration Points

### Order Service Integration

- Consumes `OrderCreated` events for stock reservation
- Publishes `StockReserved` or `StockReservationFailed` events
- Handles stock release on order cancellation

### Payment Service Integration

- Consumes `PaymentProcessed` events
- Commits reserved stock on successful payment
- Releases reserved stock on payment failure

## Error Handling

### Domain Errors

- `ErrProductNotFound`: Product not found
- `ErrInsufficientStock`: Not enough stock available
- `ErrInvalidProductID`: Invalid product identifier
- `ErrInvalidQuantity`: Invalid quantity value
- `ErrStockNotFound`: Stock information not found

## Performance Considerations

### Database Optimization

- Indexed queries for product and stock lookups
- Efficient pagination for product listings
- Optimized stock operations

### Event Processing

- Asynchronous event publishing
- Retry mechanisms for failed events
- Batch processing for high-volume scenarios

## Development

### Prerequisites

- Go 1.21+
- PostgreSQL
- Apache Kafka
- Protocol Buffers compiler

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

### Docker

```bash
# Build and run with docker-compose
docker-compose up inventory-service
```
