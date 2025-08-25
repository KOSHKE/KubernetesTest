# KubernetesTest Project Documentation

## Project Overview

**KubernetesTest** is a microservices-based order management system built with Go, React, and modern cloud-native technologies. The project demonstrates microservices architecture using gRPC, Kafka for events, PostgreSQL for data, and Redis for caching.

## System Architecture

### General Scheme
```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Frontend  │    │ API Gateway │    │   Services  │
│   (React)   │◄──►│   (Gin)     │◄──►│    (Go)     │
└─────────────┘    └─────────────┘    └─────────────┘
                          │                   │
                          ▼                   ▼
                   ┌─────────────┐    ┌─────────────┐
                   │   Kafka     │    │ PostgreSQL  │
                   │ (Messages)  │    │  (Data)     │
                   └─────────────┘    └─────────────┘
                          ▲                   │
                          │                   │
                   ┌──────┴──────┐            │
                   │   Services  │◄───────────┘
                   │ (Direct     │
                   │  Kafka)     │
                   └─────────────┘
```

### Microservices
1. **API Gateway** - HTTP API with authentication middleware (port 8080)
2. **User Service** - User management and JWT authentication (port 50051)
3. **Order Service** - Order processing and management (port 50052)
4. **Inventory Service** - Product catalog and inventory management (port 50053)
5. **Payment Service** - Payment processing and refunds (port 50054)

## Project Structure

```
KubernetesTest/
├── .cursor/                    # Cursor IDE configuration
├── .git/                      # Git repository
├── frontend/                  # React + TypeScript frontend
├── pkg/                       # READY-TO-USE packages - USE THESE!
│   ├── jwt/                   # JWT utilities and validation
│   ├── kafkaclient/           # Kafka consumer and publisher
│   ├── logger/                # Unified logging interface
│   ├── metrics/               # Prometheus metrics interface and implementation
│   └── redisclient/           # Redis client with connection pooling
├── proto/                     # Protobuf definitions
├── proto-go/                  # Generated Go files from proto
├── services/                  # Microservices
│   ├── api-gateway/          # API Gateway service
│   ├── user-service/         # User service
│   ├── order-service/        # Order service
│   ├── inventory-service/    # Inventory service
│   └── payment-service/      # Payment service
├── docker-compose.yml         # Docker Compose configuration
├── Makefile                   # Development commands
├── go.mod                     # Go module (root - single module for entire project)
├── go.sum                     # Go dependencies
├── README.md                  # Project documentation
└── .env.example              # Environment variables example
```

**IMPORTANT**: The project uses a single Go module structure with `go.mod` located only in the root directory. All services share the same module and dependencies.

**CRITICAL**: In `/pkg` directory there are READY-TO-USE implementations of common packages. ALWAYS use these instead of creating new ones or importing external packages for the same functionality.

## Ready-to-Use Packages in /pkg

### 1. Logger Package (`pkg/logger/`)
**Interface**: `Logger` with methods: `Error`, `Warn`, `Info`, `Debug`
**Implementation**: `ZapLogger` adapter for `*zap.SugaredLogger`
**Usage**: 
```go
import "github.com/kubernetestest/ecommerce-platform/pkg/logger"
// Use logger.NewZapLogger(zapLogger) to create adapter
```

### 2. Kafka Client (`pkg/kafkaclient/`)
**Consumer**: `Consumer` with worker pool optimization
**Publisher**: `Publisher` for sending messages
**Features**: 
- Worker pool for parallel processing
- Configurable buffer sizes
- Graceful shutdown
- Structured logging integration
**Usage**: Always use this instead of direct confluent-kafka-go

### 3. JWT Package (`pkg/jwt/`)
**Features**: Token generation, validation, refresh logic
**Usage**: Use for all JWT operations instead of direct golang-jwt

### 4. Redis Client (`pkg/redisclient/`)
**Features**: Connection pooling, health checks, structured logging
**Usage**: Use for all Redis operations instead of direct go-redis

### 5. Metrics Package (`pkg/metrics/`)
**Features**: Prometheus metrics interface and implementation
**Implementation**: `PrometheusMetrics` with configurable service name
**Usage**: Use for HTTP metrics and as base for service-specific metrics
**Service-Specific**: Each service has its own `internal/metrics/` package
**Reusability**: Can create service-specific metrics from existing `PrometheusMetrics` instances

## Technology Stack

### Backend (Go)
- **Language**: Go 1.25
- **Module Structure**: Single Go module (root `go.mod`) for all services
- **Framework**: Gin (HTTP), gRPC
- **ORM**: GORM with PostgreSQL driver
- **Authentication**: JWT (golang-jwt/jwt/v5) - USE pkg/jwt wrapper
- **Logging**: Zap (structured logging) - USE pkg/logger wrapper
- **Migrations**: GORM AutoMigrate
- **Hot Reload**: Air

### Frontend (React)
- **Framework**: React 18 + TypeScript
- **Bundler**: Vite
- **Styling**: CSS modules
- **HTTP Client**: Axios

### Infrastructure
- **Database**: PostgreSQL 15.4
- **Cache**: Redis 7.2-alpine - USE pkg/redisclient
- **Message Queue**: Kafka 4.0.0 (KRaft) - USE pkg/kafkaclient
- **Containerization**: Docker + Docker Compose
- **Kafka UI**: provectuslabs/kafka-ui
- **Monitoring**: Prometheus 2.48.0 + Grafana 10.2.0

### Protocols and API
- **HTTP REST**: API Gateway (Gin)
- **gRPC**: Inter-service communication
- **Protobuf**: Data serialization
- **Swagger/OpenAPI**: API documentation

## Detailed Service Structure

### User Service
```
services/user-service/
├── cmd/
│   └── main.go              # Entry point
├── internal/
│   ├── app/
│   │   ├── config.go        # Configuration
│   │   └── run.go           # Service startup
│   ├── domain/
│   │   ├── entities/        # Domain entities
│   │   └── valueobjects/    # Value objects
│   ├── infra/
│   │   ├── auth/            # JWT authentication (uses pkg/jwt)
│   │   ├── grpc/            # gRPC server
│   │   └── repository/      # GORM repository
│   └── ports/               # Interfaces (auth, repository)
├── Dockerfile                # Docker image
└── .air.toml                # Air configuration
```

### API Gateway
```
services/api-gateway/
├── cmd/
│   └── main.go              # Entry point
├── internal/
│   ├── app/
│   │   └── run.go           # Service startup
│   ├── clients/              # gRPC clients to services
│   ├── config/               # Configuration
│   ├── handlers/             # HTTP handlers (user, order, payment, inventory)
│   ├── middleware/           # HTTP middleware (auth, CORS)
│   └── pkg/                  # Internal packages
├── Dockerfile                # Docker image
└── .air.toml                # Air configuration
```

### Order Service
```
services/order-service/
├── cmd/
│   └── main.go              # Entry point
├── internal/
│   ├── app/
│   │   ├── config.go        # Configuration
│   │   └── run.go           # Service startup
│   ├── domain/
│   │   ├── errors/          # Domain errors
│   │   └── models/          # Order models
│   ├── infra/
│   │   ├── clock/           # System clock interface
│   │   ├── grpc/            # gRPC server
│   │   ├── kafka/           # Kafka integration (uses pkg/kafkaclient)
│   │   ├── productinfo/     # Product info provider
│   │   └── repository/      # GORM repository
│   └── ports/               # Interfaces
├── Dockerfile                # Docker image
└── .air.toml                # Air configuration
```

### Inventory Service
```
services/inventory-service/
├── cmd/
│   └── main.go              # Entry point
├── internal/
│   ├── app/
│   │   ├── config.go        # Configuration
│   │   └── run.go           # Service startup
│   ├── domain/
│   │   └── models/          # Product, Category, Stock models
│   ├── infra/
│   │   ├── grpc/            # gRPC server
│   │   ├── kafka/           # Kafka integration (uses pkg/kafkaclient)
│   │   └── repository/      # GORM repository
│   └── ports/               # Interfaces
├── Dockerfile                # Docker image
└── .air.toml                # Air configuration
```

### Payment Service
```
services/payment-service/
├── cmd/
│   └── main.go              # Entry point
├── internal/
│   ├── app/
│   │   ├── config.go        # Configuration
│   │   └── run.go           # Service startup
│   ├── domain/
│   │   ├── entities/        # Payment entity
│   │   ├── errors/          # Domain errors
│   │   └── valueobjects/    # Money value object
│   ├── infra/
│   │   ├── cache/           # Order totals cache
│   │   ├── grpc/            # gRPC server
│   │   ├── kafka/           # Kafka integration (uses pkg/kafkaclient)
│   │   └── processor/       # Payment processor
│   └── ports/               # Interfaces
├── Dockerfile                # Docker image
└── .air.toml                # Air configuration
```

## Configuration

### Environment Variables
The project uses environment variables for configuration. A `.env.example` file is provided as a template. Key configuration areas include:

- **Database**: PostgreSQL connection settings
- **JWT**: Authentication secrets and token TTLs
- **Redis**: Cache connection settings
- **Kafka**: Message broker configuration
- **Services**: Inter-service communication URLs

### Port Mapping
- **Frontend**: 3001:3000
- **API Gateway**: 8080:8080
- **API Gateway Metrics**: 8081:8081
- **User Service**: 50051:50051
- **User Service Metrics**: 9091:9091
- **Order Service**: 50052:50052
- **Order Service Metrics**: 9095:9095
- **Inventory Service**: 50053:50053
- **Inventory Service Metrics**: 9096:9096
- **Payment Service**: 50054:50054
- **Payment Service Metrics**: 9097:9097
- **PostgreSQL**: 5432:5432
- **Redis**: 6379:6379
- **Kafka**: 9092:9092
- **Kafka UI**: 8085:8080
- **Prometheus**: 9090:9090
- **Grafana**: 3000:3000

## API Endpoints

### Public Routes
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/refresh` - Token refresh
- `POST /api/v1/auth/logout` - User logout
- `GET /api/v1/inventory/*` - Product browsing

### Protected Routes (require JWT)
- `GET /api/v1/users/profile` - Get user profile
- `PUT /api/v1/users/profile` - Update user profile
- `POST /api/v1/orders` - Create new order
- `GET /api/v1/orders` - List user orders
- `GET /api/v1/orders/:id` - Get order details
- `POST /api/v1/payments` - Process payment
- `GET /api/v1/payments/:id` - Get payment details
- `POST /api/v1/payments/:id/refund` - Process refund

## JWT Authentication

### Tokens
- **Access Token**: Short-lived (15 minutes) for API requests
- **Refresh Token**: Long-lived (7 days) stored in Redis

### Authentication Flow
1. User logs in → receives access + refresh tokens
2. Access token used for API requests
3. When expired → automatic refresh using refresh token
4. On logout → refresh token revoked from Redis

## Events and Kafka

### Event Topics
- **order_created** - Order creation
- **payment_processed** - Payment processing
- **stock_reserved** - Stock reservation
- **stock_events** - Inventory events

### Event Flow
1. **Order Service** → `order_created` → **Inventory Service**
2. **Inventory Service** → `stock_reserved` → **Payment Service**
3. **Payment Service** → `payment_processed` → **Order Service**

### Kafka Integration Architecture
- **Direct Service Integration**: Services communicate directly with Kafka, not through API Gateway
- **Event-Driven Communication**: Asynchronous messaging between services for loose coupling
- **Service Autonomy**: Each service can independently publish and consume events
- **Scalability**: Services can scale independently based on event processing needs
- **USE pkg/kafkaclient**: All Kafka operations must use the unified client package

## Development

### Makefile Commands
```bash
make help              # Show available commands
make proto             # Generate protobuf stubs
make proto-clean       # Clean generated files
make dev-up            # Start dev environment
make dev-rebuild       # Rebuild and start
make dev-down          # Stop dev environment
make fmt               # Format Go code
```

### Hot Reload with Air
Each service supports hot reload through Air:
```bash
# Navigate to service directory
cd services/user-service

# Start with hot reload
air
```

### Protobuf Generation
```bash
# Generate Go files from proto
make proto

# Clean generated files
make proto-clean
```

## Monitoring and Logging

### Health Checks
- All services have health check endpoints
- Docker Compose uses health checks for dependencies
- Infrastructure services (PostgreSQL, Redis, Kafka) have health checks with `service_healthy` condition
- Go services wait for healthy infrastructure before starting
- Prometheus waits for all Go services to start before collecting metrics

### Logging
- **Structured logging** through Zap - USE pkg/logger wrapper
- **Correlation ID** for request tracking
- **Request ID** for unique identification

### Monitoring
- **Kafka UI**: http://localhost:8085
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)
- **Health endpoints**: `/health` on each service
- **Metrics endpoints**: `/metrics` on each service
- **Structured logs** in JSON format

## Metrics Architecture

### Overview
The project implements a comprehensive metrics system following Prometheus best practices with **consistent labeling** across all metric types. This ensures unified analysis, simplified dashboard creation, and better correlation between business and technical metrics.

### Metrics Structure

#### 1. **HTTP/gRPC Metrics (pkg/metrics)**
- **Request Counters**: `http_requests_total{service, method, endpoint, status_code}`
- **Request Duration**: `http_request_duration_seconds{service, method, endpoint}`
- **Standardized across all services** for technical monitoring

#### 2. **Business Metrics (Service-Specific)**
Each service implements domain-specific metrics with consistent labeling:
- **Success Metrics**: `{service, method, status}` - e.g., `payment_succeeded_total{service="payment-service", method="CREDIT_CARD", status="success"}`
- **Failure Metrics**: `{service, method, failure_reason}` - e.g., `payment_failed_total{service="payment-service", method="unknown", failure_reason="insufficient_funds"}`
- **Duration Metrics**: `{service, method}` - e.g., `payment_processing_duration_seconds{service="payment-service", method="CREDIT_CARD"}`

### Label Consistency Benefits

#### **Unified Label Structure**
- **`service`** - consistent across all metrics (e.g., "payment-service", "user-service")
- **`method`** - business method (CREDIT_CARD, DEBIT_CARD) matching HTTP method concept
- **`status`** - success/failure status (similar to HTTP status_code)
- **`failure_reason`** - specific failure reasons for business logic

#### **Cross-Metric Analysis**
```promql
# Compare business success rate vs HTTP success rate
rate(payment_succeeded_total{service="payment-service"}[5m]) vs
rate(http_requests_total{service="payment-service", status_code="200"}[5m])

# Analyze performance by method
rate(payment_processing_duration_seconds_sum{service="payment-service", method="CREDIT_CARD"}[5m]) /
rate(payment_processing_duration_seconds_count{service="payment-service", method="CREDIT_CARD"}[5m])
```

### Metrics Implementation

#### **Service-Specific Metrics Package**
Each service has its own `internal/metrics/` package:
```go
// Example: Payment Service Metrics
type PaymentMetrics interface {
    PaymentSucceeded(method string)
    PaymentFailed(reason string)
    PaymentProcessingDuration(duration time.Duration, method string)
    metrics.Metrics // Inherits HTTP metrics from pkg/metrics
}
```

#### **Prometheus Integration**
- **CounterVec** for success/failure metrics with proper labels
- **HistogramVec** for duration metrics with configurable buckets
- **Automatic registration** using `promauto` for zero-config setup

### Grafana Dashboards

#### **Dashboard Structure**
Each service has a dedicated dashboard showing:
1. **Business Rate Metrics** - success/failure rates per second
2. **Cumulative Counts** - total successful/failed operations
3. **Performance Metrics** - average duration and percentiles
4. **Service Health** - Prometheus `up` metric
5. **Comparison Views** - pie charts for success vs failure ratios

#### **Dashboard Configuration**
- **Refresh Rate**: 5-10 seconds for real-time monitoring
- **Time Range**: 1 hour default with zoom capabilities
- **Panel Types**: Stat panels for single values, timeseries for trends
- **PromQL Queries**: Optimized for consistent label usage

### Best Practices Applied

#### **Label Cardinality Control**
- **Low Cardinality**: `method` has limited values (CREDIT_CARD, DEBIT_CARD, etc.)
- **Meaningful Labels**: `status` and `failure_reason` provide business context
- **Consistent Naming**: Matches `pkg/metrics` HTTP metric patterns
- **Avoiding Duplication**: No redundant labels when context is already clear

#### **Metrics Naming Convention**
- **Descriptive Names**: Clear purpose and data representation
- **Standardized Units**: Time in seconds, sizes in bytes
- **Consistent Format**: Lowercase with underscores
- **Help Text**: Comprehensive descriptions for each metric

### Benefits of Consistent Metrics Architecture

1. **Unified Analysis**: Easy correlation between business and HTTP metrics
2. **Simplified Dashboards**: Same label names across different metric types
3. **Better Aggregation**: Group by `method` across all metric types
4. **Reduced Learning Curve**: Developers familiar with HTTP metrics understand business metrics
5. **Easier Troubleshooting**: Trace issues from business logic to HTTP layer
6. **Scalable Monitoring**: Consistent pattern for future services
7. **Effective Alerting**: Reliable threshold setting and anomaly detection

### Metrics Documentation
Each service includes comprehensive metrics documentation:
- **METRICS.md** files with examples and PromQL queries
- **Label descriptions** and usage guidelines
- **Cross-metric analysis** examples
- **Best practices** for dashboard creation

## Security

### JWT
- Token validation on all protected routes - USE pkg/jwt
- Refresh token storage in Redis with revocation capability
- Configurable token lifetime

### CORS
- CORS configuration for frontend origins
- Secure default settings

### Validation
- Input data validation
- User input sanitization
- SQL injection protection through GORM

## Deployment

### Docker
- Multi-stage Dockerfiles for size optimization
- Health checks for all services
- Volume mounts for hot reload in development

## Development Principles

### Go Best Practices
- Following Effective Go
- Idiomatic Go code
- Error handling through error return
- Using context.Context for timeouts

### Architectural Principles
- **Clean Architecture** / Layered approach
- **Dependency Injection** through interfaces
- **Separation of Concerns** between layers
- **Interface Segregation** for flexibility

### Code Style
- Small functions with clear names
- No "magic numbers"
- GoDoc style comments
- Structured logging

### Package Usage Rules
1. **ALWAYS use pkg/logger** instead of direct zap
2. **ALWAYS use pkg/kafkaclient** instead of direct confluent-kafka-go
3. **ALWAYS use pkg/jwt** instead of direct golang-jwt
4. **ALWAYS use pkg/metrics** instead of direct prometheus client
5. **ALWAYS use pkg/redisclient** instead of direct go-redis
6. **NEVER create new packages** for functionality that already exists in /pkg
7. **NEVER import external packages** for functionality available in /pkg

## Common Patterns

### Service Initialization
```go
// Always use pkg packages
logger := logger.NewZapLogger(zapLogger)
kafkaConsumer := kafkaclient.NewConsumer(config)
redisClient := redisclient.NewClient(config)
jwtService := jwt.NewService(config)

// For metrics, use service-specific package
metricsInstance := usermetrics.NewUserMetrics() // in user-service
```

### Error Handling
```go
// Use domain-specific errors
if err != nil {
    return fmt.Errorf("failed to process order: %w", err)
}
```

### Logging
```go
// Use structured logging through pkg/logger
logger.Error("failed to connect to database", "error", err, "host", host)
```

## Conclusion

The KubernetesTest project represents a modern microservices architecture using Go, demonstrating best practices for developing cloud-native applications. The architecture is ready for Kubernetes deployment and can be easily extended for production use.

**REMEMBER**: Always use the ready-to-use packages in `/pkg` directory. They provide unified interfaces, optimized implementations, and consistent behavior across all services.