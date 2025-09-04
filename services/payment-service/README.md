# Payment Service

## Overview

The Payment Service is a core microservice within the e-commerce platform responsible for processing payments and managing payment-related events. It handles payment processing requests, publishes payment events, and consumes stock reservation events to trigger payment processing.

**Role in the Platform:**
- Payment processing and validation
- Payment event publishing (PaymentProcessed)
- Stock event consumption (StockReserved)
- Integration with order and inventory services

## Architecture

### Layered Architecture

The service follows Domain-Driven Design (DDD) principles with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                Infrastructure Layer                         │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐          │
│  │   gRPC      │ │   Kafka     │ │   Kafka     │          │
│  │   Server    │ │ Publisher   │ │ Consumer    │          │
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
│  │  Entities   │ │   Ports     │ │ Value      │          │
│  │ (Payment)   │ │(Interfaces) │ │ Objects    │          │
│  └─────────────┘ └─────────────┘ └─────────────┘          │
└─────────────────────────────────────────────────────────────┘
```

### Technology Stack

- **Language**: Go 1.21+
- **Communication**: gRPC with Protocol Buffers
- **Event Streaming**: Apache Kafka (publishing and consuming)
- **Metrics**: Prometheus-compatible metrics with HTTP endpoint
- **Logging**: Structured logging with Zap
- **Configuration**: Environment-based configuration management

### Dependency Organization

```
Payment Service
├── Domain Layer
│   ├── Entities (Payment)
│   ├── Value Objects (PaymentMethod, PaymentStatus)
│   └── Ports (Publisher interfaces)
├── Application Layer
│   ├── Use Cases (ProcessPayment)
│   ├── Application Services (PaymentApplicationService, StockEventService)
│   └── DTOs (Request/Response models)
├── Infrastructure Layer
│   ├── gRPC Server
│   ├── Kafka Publisher (PaymentProcessed)
│   ├── Kafka Consumer (StockReserved)
│   └── Event Handlers
└── Cross-cutting Concerns
    ├── Metrics (Prometheus)
    ├── Logging (Zap)
    └── Error Handling (domain-specific errors)
```

## Data Flow

### Payment Processing Flow

```
Client Request → gRPC Server → Application Service → Use Case → Publisher → Kafka
                                                                           ↓
Response ← gRPC Server ← Application Service ← Use Case ← Publisher ← Kafka
```

### Event-Driven Flow

```
StockReserved Event → Kafka Consumer → StockEventService → ProcessPaymentUseCase → PaymentProcessed Event
```

### Event Publishing Flow

```
Payment Processed → Payment Service (PaymentProcessed event) → Order Service
```

## Components

### Payment Entity

**Location**: `internal/domain/entities/payment.go`

The Payment entity represents the payment aggregate root:

**Key Methods**:
- `NewPayment(orderID, userID, amount, method)` - Create new payment
- `CanBeProcessed()` - Check if payment can be processed
- `ID` - Get payment ID
- `OrderID` - Get associated order ID
- `UserID` - Get user ID
- `Amount` - Get payment amount
- `Status` - Get payment status
- `Method` - Get payment method
- `TransactionID` - Get transaction ID

### Value Objects

**Location**: `internal/domain/valueobjects/`

1. **PaymentMethod** - Payment method enumeration
   - `PaymentMethodCreditCard` - Credit card payment
   - `String()` - Get string representation

2. **PaymentStatus** - Payment status enumeration
   - `PaymentStatusPending` - Payment is pending
   - `PaymentStatusCompleted` - Payment completed successfully
   - `PaymentStatusFailed` - Payment failed
   - `IsPending()` - Check if status is pending
   - `IsCompleted()` - Check if status is completed
   - `IsFailed()` - Check if status is failed

### Use Cases

**Location**: `internal/application/usecases/`

1. **ProcessPaymentUseCase**
   - Processes payment requests
   - Simulates payment processing (50% success rate)
   - Publishes PaymentProcessed events
   - Handles payment status updates

**Key Methods**:
- `Execute(ctx, orderID, userID, amount, method)` - Process payment

### Application Services

**Location**: `internal/application/services/`

1. **PaymentApplicationService**
   - Orchestrates payment operations
   - Provides unified interface for gRPC server
   - Handles DTO conversion

2. **StockEventService**
   - Processes StockReserved events
   - Converts events to payment processing requests
   - Triggers payment processing

**Key Methods**:
- `ProcessPayment(ctx, req)` - Process payment request
- `ProcessStockReserved(ctx, evt)` - Process stock reserved event

### gRPC Server

**Location**: `internal/infra/grpc/pb_payment_server.go`

- Protocol Buffer-based gRPC server
- Health check service integration
- Reflection service (development mode)
- Error mapping to gRPC status codes
- Request/response DTO mapping

**Services**:
- `ProcessPayment` - Process payment request

### Kafka Publisher

**Location**: `internal/infra/publisher/payment_processed_publisher.go`

- Type-safe Kafka publisher for PaymentProcessed events
- Uses protobuf for serialization
- Configurable client settings

**Key Methods**:
- `PublishPaymentProcessed(ctx, evt)` - Publish payment processed event

### Kafka Consumer

**Location**: `internal/infra/consumer/stock_reserved_consumer.go`

- Type-safe Kafka consumer for StockReserved events
- Uses protobuf for deserialization
- Configurable consumer settings

**Key Methods**:
- `Run(ctx, topics)` - Start consuming events

### Event Handlers

**Location**: `internal/infra/consumer/handlers/stock_reserved_handler.go`

- Handles StockReserved events
- Delegates processing to StockEventService
- Provides error handling and logging

**Key Methods**:
- `Handle(ctx, evt)` - Process stock reserved event

### Metrics

**Location**: `internal/metrics/payment_metrics.go`

- Prometheus-compatible metrics endpoint
- Payment-specific business metrics
- HTTP metrics (request count, duration, errors)
- Separate port for metrics collection

**Business Metrics**:
- `payment_succeeded_total` - Successful payments
- `payment_failed_total` - Failed payments by reason
- Entity events with payment entity type and actions

## Lifecycle

### Application Initialization

```
main() → Run() → initialize() → start() → waitForShutdown()
```

**Initialize Phase**:
1. **Infrastructure**: Metrics server, gRPC server
2. **Business Logic**: Application services, use cases
3. **Kafka Components**: Publishers and consumers
4. **gRPC Server**: Server setup, service registration, health checks

**Start Phase**:
- Start gRPC server on configured port
- Start metrics HTTP server in background
- Start Kafka consumers in background
- Application ready to serve requests

### Graceful Shutdown

**Shutdown Sequence**:
1. **Signal Handling**: OS interrupt signals (SIGTERM, SIGINT, SIGHUP)
2. **gRPC Server**: Graceful stop with 5-second timeout
3. **Kafka Components**: Close publishers and consumers
4. **Resource Cleanup**: Close all closers
5. **Context Cancellation**: Cancel all background operations

**Shutdown Timeout**:
- gRPC graceful shutdown: 5 seconds
- Force stop if graceful shutdown fails

## Configuration Management

**Environment Variables**:
- `PAYMENT_SERVICE_PORT`: gRPC server port (default: 50054)
- `PAYMENT_SERVICE_METRICS_PORT`: Metrics HTTP port (default: 9097)
- `KAFKA_BROKERS`: Kafka broker addresses (comma-separated)
- `REDIS_URL`: Redis connection URL (default: redis://redis:6379)
- `REDIS_ADDR`: Redis address (alternative to REDIS_URL)
- `REDIS_DB`: Redis database number (default: 0)
- `REDIS_PASSWORD`: Redis password
- `REDIS_POOL_SIZE`: Redis connection pool size (default: 10)
- `REDIS_MIN_IDLE_CONNS`: Redis minimum idle connections (default: 5)
- `REDIS_DIAL_TIMEOUT`: Redis dial timeout (default: 5s)
- `REDIS_READ_TIMEOUT`: Redis read timeout (default: 3s)
- `REDIS_WRITE_TIMEOUT`: Redis write timeout (default: 3s)
- `PAYMENT_PROCESS_TIMEOUT`: Payment processing timeout (default: 30s)
- `KAFKA_PUBLISH_TIMEOUT`: Kafka publish timeout (default: 10s)
- `PAYMENT_MAX_RETRIES`: Maximum payment retries (default: 3)
- `PAYMENT_RETRY_DELAY`: Payment retry delay (default: 1s)
- `PAYMENT_DEFAULT_CURRENCY`: Default currency (default: USD)
- `PAYMENT_SUPPORTED_METHODS`: Supported payment methods (default: CREDIT_CARD,DEBIT_CARD,BANK_TRANSFER)
- `ORDER_TOTAL_TTL`: Order total TTL (default: 30m)

**Validation**:
- Required Redis configuration
- Positive timeout values
- Valid currency codes
- Supported payment methods

## Development

### Prerequisites

- Go 1.21+
- Apache Kafka
- Redis
- Protocol Buffers compiler

### Local Development

```bash
# Set environment variables
export KAFKA_BROKERS="localhost:9092"
export REDIS_URL="redis://localhost:6379"
export PAYMENT_SERVICE_PORT="50054"
export PAYMENT_SERVICE_METRICS_PORT="9097"

# Run service
go run cmd/main.go
```

### Docker

```bash
# Build and run with docker-compose
docker-compose up payment-service
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

**Business Metrics**:
- `payment_succeeded_total` - Successful payments
- `payment_failed_total` - Failed payments by reason
- Entity events with payment entity type and actions

**HTTP Metrics**:
- Request count and duration
- Error rates and status codes
- Response time percentiles

### Health Checks

- gRPC health check service
- Kafka connectivity verification
- Redis storage availability

### Logging

- Structured JSON logging
- Request correlation IDs
- Error context and stack traces
- Performance metrics integration

## Integration Points

### Order Service Integration

- **Consumes**: OrderCreated events (via StockReserved)
- **Publishes**: PaymentProcessed events
- **Purpose**: Process payments when stock is reserved

### Inventory Service Integration

- **Consumes**: StockReserved events
- **Purpose**: Trigger payment processing when inventory is confirmed

### Event Flow

1. **Order Service** → OrderCreated event → **Inventory Service**
2. **Inventory Service** → StockReserved event → **Payment Service**
3. **Payment Service** → PaymentProcessed event → **Order Service**

## Error Handling

**Domain Errors**:
- `ErrPaymentAlreadyProcessed` - Payment already processed
- `ErrPaymentProcessingFailed` - Payment processing failed
- `ErrPaymentNotFound` - Payment not found
- `ErrInvalidPaymentMethod` - Invalid payment method
- `ErrInvalidPaymentAmount` - Invalid payment amount
- `ErrPaymentTimeout` - Payment processing timeout
- `ErrPaymentRetryExceeded` - Maximum retries exceeded

**Error Recovery**:
- Automatic retry with exponential backoff
- Dead letter queue for failed events
- Circuit breaker pattern for external services
- Graceful degradation on service unavailability

## Security Considerations

- Payment data validation and sanitization
- Secure event publishing and consuming
- Input validation for all payment requests
- Rate limiting for payment processing
- Audit logging for payment operations
