# User Service

## Overview

The User Service is a core microservice within the e-commerce platform responsible for user management, authentication, and authorization. It provides secure user registration, login/logout functionality, and token-based authentication using JWT (JSON Web Tokens) with Redis-based token storage.

**Role in the Platform:**
- Centralized user identity management
- Authentication and authorization services
- User profile management
- Integration point for other services requiring user validation

## Architecture

### Layered Architecture

The service follows Domain-Driven Design (DDD) principles with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                Infrastructure Layer                         │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐          │
│  │   gRPC      │ │ Repository  │ │    Auth    │          │
│  │   Server    │ │   (GORM)    │ │  Service   │          │
│  │             │ │ Migration   │ │ Redis      │          │
│  └─────────────┘ └─────────────┘ └─────────────┘          │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                Application Layer                            │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐          │
│  │   Use       │ │ Application │ │     DTOs    │          │
│  │   Cases     │ │  Service    │ │             │          │
│  └─────────────┘ └─────────────┘ └─────────────┘          │
└─────────────────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────────────────┐
│                  Domain Layer                               │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐          │
│  │  Entities   │ │   Ports     │ │ Value      │          │
│  │   (User)    │ │(Interfaces) │ │ Objects    │          │
│  └─────────────┘ └─────────────┘ └─────────────┘          │
└─────────────────────────────────────────────────────────────┘
```

### Technology Stack

- **Language**: Go 1.21+
- **Communication**: gRPC with Protocol Buffers
- **Database**: PostgreSQL with GORM ORM
- **Authentication**: JWT (Access + Refresh tokens)
- **Token Storage**: Redis
- **Metrics**: Prometheus-compatible metrics with HTTP endpoint
- **Logging**: Structured logging with Zap
- **Configuration**: Environment-based configuration management

### Dependency Organization

```
User Service
├── Domain Layer
│   ├── Entities (User aggregate)
│   ├── Value Objects (Email, Password, Name, Phone, TokenPair)
│   └── Ports (Repository, Auth Service interfaces)
├── Application Layer
│   ├── Use Cases (Register, Login, GetUser, RefreshToken, Logout)
│   ├── Application Service (orchestrates use cases)
│   └── DTOs (Request/Response models)
├── Infrastructure Layer
│   ├── Repository (GORM implementation)
│   ├── Auth Service (JWT implementation with Redis)
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

### Authentication Flow

```
1. Login Request → Validate Credentials → Generate JWT Tokens
2. Store Refresh Token in Redis → Return Access + Refresh Tokens
3. Client uses Access Token for subsequent requests
4. Token Refresh: Validate Refresh Token → Generate New Token Pair
5. Logout: Revoke Refresh Token from Redis
```

### Metrics and Logging Integration

- **Metrics**: Collected at use case level (user creation, login success/failure)
- **Logging**: Structured logging throughout the request lifecycle
- **Health Checks**: gRPC health check service for service discovery

## Components

### User Entity

**Location**: `internal/domain/entities/user.go`

The User entity represents the user aggregate root:

**Key Methods**:
- `NewUser(id, email, password, firstName, lastName, phone)` - Create new user
- `ID()` - Get user ID
- `Email()` - Get email value object
- `Password()` - Get password value object
- `FirstName()` - Get first name value object
- `LastName()` - Get last name value object
- `Phone()` - Get phone value object
- `CreatedAt()` - Get creation timestamp
- `UpdatedAt()` - Get update timestamp

### Value Objects

**Location**: `internal/domain/valueobjects/`

1. **Email** - Email address value object
   - `NewEmail(email)` - Create email with validation
   - `Value()` - Get email value
   - `Equals(other)` - Compare emails
   - Automatic normalization (lowercase, trimmed)

2. **Password** - Password value object with bcrypt hashing
   - `NewPasswordFromPlain(plainPassword)` - Create from plain text with hashing
   - `NewPassword(hashedPassword)` - Create from hashed password
   - `Verify(plainPassword)` - Verify password against hash
   - `HashedValue()` - Get hashed password

3. **Name** - Name value object
   - `NewName(name)` - Create name with validation
   - `Value()` - Get name value
   - `Equals(other)` - Compare names
   - Automatic trimming

4. **Phone** - Phone number value object
   - `NewPhone(phone)` - Create phone with validation
   - `Value()` - Get phone value
   - `Equals(other)` - Compare phones

5. **TokenPair** - Authentication token pair
   - `NewTokenPair(accessToken, refreshToken, expiresIn)` - Create token pair
   - `IsExpired()` - Check if access token is expired
   - `GetExpiresInSeconds()` - Get seconds until expiration

### User Repository (GORM)

**Location**: `internal/infra/repository/gorm_user_repository.go`

- Implements `UserRepository` interface
- Uses GORM for database operations
- Separates domain entities from database records
- Supports transactions via `WithTx()` method
- Handles CRUD operations for user management

**Key Methods**:
- `Create(ctx, user)` - Create new user
- `GetByID(ctx, id)` - Retrieve user by ID
- `GetByEmail(ctx, email)` - Retrieve user by email
- `ExistsByEmail(ctx, email)` - Check email uniqueness
- `ExistsByID(ctx, id)` - Check user existence by ID
- `Update(ctx, user)` - Update user information
- `Delete(ctx, id)` - Delete user

### Authentication Service (JWT)

**Location**: `internal/infra/auth/auth_jwt.go`

- JWT-based authentication with access and refresh tokens
- Redis-based token storage for refresh tokens
- Configurable token TTL (access: 15min, refresh: 7 days)
- Secure token generation and validation

**Key Methods**:
- `GenerateTokenPair(userID, email)` - Generate new token pair
- `StoreRefreshToken(ctx, refreshToken, userID)` - Store refresh token in Redis
- `ValidateRefreshToken(tokenString)` - Validate refresh token
- `RefreshAccessToken(refreshToken)` - Generate new token pair from refresh token
- `RevokeRefreshToken(ctx, refreshToken)` - Revoke specific refresh token
- `RevokeAllUserTokens(ctx, userID)` - Revoke all user tokens

**Features**:
- Access token for API authentication
- Refresh token for token renewal
- Redis-based token blacklisting
- Configurable secrets and expiration times
- Token validation with Redis storage check

### Redis Token Storage

**Location**: `internal/infra/auth/token_storage_redis.go`

- Redis-based implementation of token storage
- Handles refresh token persistence and validation
- Supports token revocation and cleanup

**Key Methods**:
- `StoreToken(ctx, token, userID, ttl)` - Store token with TTL
- `IsTokenValid(ctx, token)` - Check if token exists and is valid
- `RevokeToken(ctx, token)` - Remove specific token
- `RevokeAllUserTokens(ctx, userID)` - Remove all user tokens

### Use Cases

**Location**: `internal/application/usecases/`

1. **RegisterUserUseCase**
   - Validates user input (email, password, names, phone)
   - Checks email uniqueness
   - Creates user entity with value objects
   - Records metrics for successful registration

2. **LoginUserUseCase**
   - Validates user credentials
   - Generates JWT token pair
   - Stores refresh token in Redis
   - Records login success/failure metrics

3. **GetUserUseCase**
   - Retrieves user information by ID
   - Returns user profile data

4. **RefreshTokenUseCase**
   - Validates refresh token
   - Generates new token pair
   - Updates token storage

5. **LogoutUseCase**
   - Revokes refresh token from Redis
   - Invalidates user session

### Application Service

**Location**: `internal/application/services/user_application_service.go`

- Orchestrates user operations
- Provides unified interface for gRPC server
- Handles DTO conversion and validation
- Coordinates use cases

**Key Methods**:
- `RegisterUser(ctx, req)` - Register new user
- `LoginUser(ctx, req)` - Authenticate user
- `GetUser(ctx, req)` - Get user information
- `RefreshToken(ctx, req)` - Refresh authentication tokens
- `Logout(ctx, req)` - Logout user

### gRPC Server

**Location**: `internal/infra/grpc/`

- Protocol Buffer-based gRPC server
- Health check service integration
- Reflection service (development mode)
- Error mapping to gRPC status codes
- Request/response DTO mapping

**Services**:
- `Register` - User registration
- `Login` - User authentication
- `GetUser` - Retrieve user profile
- `RefreshToken` - Token renewal
- `Logout` - User logout

### Migration Service

**Location**: `internal/infra/migration/`

- Database schema initialization
- Automatic migration execution on startup
- Ensures database consistency

### Metrics

**Location**: `internal/metrics/`

- Prometheus-compatible metrics endpoint
- User-specific business metrics
- HTTP metrics (request count, duration, errors)
- Separate port for metrics collection

**Business Metrics**:
- `user_created_total` - Total user registrations
- `user_login_success_total` - Successful logins
- `user_login_failed_total` - Failed logins by reason
- Entity events with user entity type and actions

## Lifecycle

### Application Initialization

```
main() → Run() → initialize() → start() → waitForShutdown()
```

**Initialize Phase**:
1. **Infrastructure**: Database connection, migrations, metrics server
2. **Business Logic**: Repository, auth service, use cases, application service
3. **gRPC Server**: Server setup, service registration, health checks

**Start Phase**:
- Start gRPC server on configured port
- Start metrics HTTP server in background
- Application ready to serve requests

### Graceful Shutdown

**Shutdown Sequence**:
1. **Signal Handling**: OS interrupt signals (SIGTERM, SIGINT, SIGHUP)
2. **gRPC Server**: Graceful stop with 5-second timeout
3. **Resource Cleanup**: Close database connections, Redis connections
4. **Context Cancellation**: Cancel all background operations

**Shutdown Timeout**:
- gRPC graceful shutdown: 5 seconds
- Force stop if graceful shutdown fails

## Extensibility & Design Patterns

### Domain-Driven Design (DDD)

- **Domain Layer**: Core business logic and entities
- **Application Layer**: Use cases and orchestration
- **Infrastructure Layer**: External concerns (database, gRPC)

### SOLID Principles

- **Single Responsibility**: Each use case handles one business operation
- **Open/Closed**: Extensible through interface implementations
- **Liskov Substitution**: Repository and auth service interfaces
- **Interface Segregation**: Focused interfaces for different concerns
- **Dependency Inversion**: High-level modules depend on abstractions

### Extension Points

**New Use Cases**:
1. Implement use case interface
2. Add to application service
3. Register in gRPC server
4. Add corresponding metrics

**New Authentication Methods**:
1. Implement `AuthService` interface
2. Configure in application startup
3. Update configuration structure

**Additional Repositories**:
1. Implement repository interface
2. Register in dependency injection
3. Use in appropriate use cases

**New Validation Rules:**
1. Add custom validation functions to validator instance
2. Use custom tags in DTO structs
3. Test validation logic in isolation
4. Update error handling for new validation failures

**Example Custom Validation:**
```go
// Custom validation function
func validatePhone(fl validator.FieldLevel) bool {
    phone := fl.Field().String()
    return regexp.MustCompile(`^\+?[1-9]\d{1,14}$`).MatchString(phone)
}

// Register in validator
validator.RegisterValidation("phone", validatePhone)

// Use in DTO
type UserRequest struct {
    Phone string `json:"phone" validate:"required,phone"`
}
```

### Configuration Management

**Environment Variables**:
- `USER_SERVICE_PORT`: gRPC server port (default: 50051)
- `USER_SERVICE_METRICS_PORT`: Metrics HTTP port (default: 9090)
- `JWT_ACCESS_SECRET`: Access token secret key (required)
- `JWT_REFRESH_SECRET`: Refresh token secret key (required)
- `ACCESS_TOKEN_TTL`: Access token lifetime (default: 15m)
- `REFRESH_TOKEN_TTL`: Refresh token lifetime (default: 168h)
- `STORAGE_URL`: Redis connection URL (default: redis://localhost:6379)
- `STORAGE_TIMEOUT`: Redis operation timeout (default: 5s)
- `USER_MAX_PASSWORD_LENGTH`: Maximum password length (default: 128)
- `USER_MIN_PASSWORD_LENGTH`: Minimum password length (default: 8)
- `USER_MAX_NAME_LENGTH`: Maximum name length (default: 50)
- `USER_MAX_PHONE_LENGTH`: Maximum phone length (default: 50)
- `USER_MAX_EMAIL_LENGTH`: Maximum email length (default: 255)

**Validation**:
- Required secrets and timeouts
- Positive duration values
- Logical constraints (max > min lengths)
- JWT secrets are mandatory

## Development

### Prerequisites

- Go 1.21+
- PostgreSQL
- Redis
- Protocol Buffers compiler

### Local Development

```bash
# Set environment variables
export JWT_ACCESS_SECRET="your-access-secret"
export JWT_REFRESH_SECRET="your-refresh-secret"
export DATABASE_URL="postgres://user:pass@localhost:5432/users"
export STORAGE_URL="redis://localhost:6379"
export USER_SERVICE_PORT="50051"
export USER_SERVICE_METRICS_PORT="9090"

# Run service
go run cmd/main.go
```

### Docker

```bash
# Build and run with docker-compose
docker-compose up user-service
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
- `user_created_total` - Total user registrations
- `user_login_success_total` - Successful logins
- `user_login_failed_total` - Failed logins by reason
- Entity events with user entity type and actions

**HTTP Metrics**:
- Request count and duration
- Error rates and status codes
- Response time percentiles

### Health Checks

- gRPC health check service
- Database connectivity verification
- Redis storage availability

### Logging

- Structured JSON logging
- Request correlation IDs
- Error context and stack traces
- Performance metrics integration
