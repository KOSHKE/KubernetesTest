# Kubernetes Test Project

## Project Overview

This project represents a microservices architecture for an order management system using Kubernetes, Docker, Go, and React.

## Architecture

### Services

1. **API Gateway** - Single entry point for all requests
2. **User Service** - User management and authentication
3. **Order Service** - Order management
4. **Inventory Service** - Product and inventory management
5. **Payment Service** - Payment processing

### Infrastructure

- **PostgreSQL** - Main database
- **Redis** - Cache and refresh token storage with SOLID architecture
- **Kafka** - Asynchronous communication between services
- **Frontend** - React application with TypeScript

## JWT Authentication

### Implemented Components

#### 1. Authentication Libraries

- **`pkg/jwt/`** - JWT token management (generation, validation, refresh)
- **`pkg/redisclient/`** - Storage and management of refresh tokens in Redis

#### 2. User Service

- **JWT Manager** - Generation of access and refresh tokens
- **Redis Client** - Storage of refresh tokens with revocation capability
- **Auth Service** - Interface for authentication
- **Updated methods**:
  - `LoginUser` - returns token pair
  - `RefreshToken` - refreshes access token
  - `Logout` - revokes refresh token

#### 3. API Gateway

- **Auth Middleware** - JWT token validation for protected routes
- **Updated routes**:
  - `/api/v1/auth/*` - public authentication routes
  - `/api/v1/users/*`, `/api/v1/orders/*`, `/api/v1/payments/*` - protected routes
  - `/api/v1/inventory/*` - public product routes

#### 4. Frontend

- **Auth Service** - Token management in localStorage
- **Login Component** - Login form
- **Automatic token refresh** - Axios interceptor for automatic access token refresh
- **Protected routes** - Redirect to login for unauthenticated users

### Configuration

#### Environment Variables

```bash
# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-in-production
JWT_REFRESH_SECRET=your-super-secret-refresh-key-change-in-production
ACCESS_TOKEN_TTL=15m
REFRESH_TOKEN_TTL=168h

# Redis
REDIS_URL=redis:6379
```

#### Docker Compose

Updated to support JWT environment variables in user-service and api-gateway.

### Security

- **Access tokens** - short-lived (15 minutes)
- **Refresh tokens** - long-lived (7 days)
- **Redis storage** - ability to revoke tokens
- **Automatic refresh** - transparent token refresh on frontend
- **Protected routes** - all order and payment operations require authentication

## Redis Architecture

### SOLID Principles Implementation

#### Single Responsibility Principle
- **`Client`** - базовое подключение к Redis
- **`TokenStorage`** - управление JWT токенами
- **`Cache`** - общие операции кэширования
- **`List`** - операции со списками

#### Interface Segregation Principle
```go
// Используйте только нужные интерфейсы
tokenStorage := redisclient.NewTokenStorage(client)
cache := redisclient.NewCache(client)
list := redisclient.NewList(client)
```

#### Dependency Inversion Principle
- Все операции через интерфейсы
- Легкое тестирование с моками
- Гибкая замена реализаций

### Performance Optimizations

#### Efficient Token Revocation
- **Было**: Сканирование всех ключей `refresh_token:*`
- **Стало**: Индекс пользователя `user_tokens:{userID}` + pipeline операции

#### Connection Pooling
- Настраиваемый размер пула
- Минимальные и максимальные соединения
- Retry логика с таймаутами

#### Unified Logging
- **Интеграция с существующим логгером** - используйте ваш Zap логгер
- **Консистентность логов** - одинаковый формат во всех компонентах
- **Correlation ID** - отслеживание запросов через сервисы
- **Структурированное логирование** - JSON формат для анализа

### Usage Examples

#### Basic Client
```go
client := redisclient.NewSimpleClient("localhost:6379", "", 0)
defer client.Close()
```

#### Advanced Configuration
```go
cfg := &redisclient.Config{
    PoolSize:     20,
    MinIdleConns: 10,
    MaxRetries:   5,
    DialTimeout:  10 * time.Second,
}
client := redisclient.NewClient(cfg)
```

#### Token Operations
```go
tokenStorage := redisclient.NewTokenStorage(client)
err := tokenStorage.StoreRefreshToken(ctx, token, userID, expiration)
```

#### Cache Operations
```go
cache := redisclient.NewCache(client)
err := cache.Set(ctx, "key", value, time.Hour)
```

#### Logging Integration
```go
// Используйте ваш существующий логгер
import "go.uber.org/zap"

logger := zap.NewProduction() // ваш существующий логгер

// Создайте Redis клиент с логгером
redisClient := redisclient.NewClientWithLogger(
    "localhost:6379", "", 0,
    redisclient.NewZapLoggerAdapter(logger),
)

// Теперь все Redis операции логируются через ваш логгер
tokenStorage := redisclient.NewTokenStorage(redisClient)
```

### API Endpoints

#### Public Routes
- `POST /api/v1/auth/register` - Registration
- `POST /api/v1/auth/login` - Login
- `POST /api/v1/auth/refresh` - Token refresh
- `POST /api/v1/auth/logout` - Logout
- `GET /api/v1/inventory/*` - Product browsing

#### Protected Routes
- `GET /api/v1/users/profile` - User profile
- `PUT /api/v1/users/profile` - Profile update
- `POST /api/v1/orders` - Create order
- `GET /api/v1/orders` - User orders list
- `GET /api/v1/orders/:id` - Order details
- `POST /api/v1/payments` - Process payment
- `GET /api/v1/payments/:id` - Payment details
- `POST /api/v1/payments/:id/refund` - Payment refund

## Project Launch

### Requirements

- Docker and Docker Compose
- Node.js 18+ (for frontend)

### Commands

```bash
# Start all services
docker-compose up -d

# Stop
docker-compose down

# View logs
docker-compose logs -f [service-name]
```

### Service Access

- **Frontend**: http://localhost:3001
- **API Gateway**: http://localhost:8080
- **Kafka UI**: http://localhost:8085
- **PostgreSQL**: localhost:5432
- **Redis**: localhost:6379

## Development

### Project Structure

```
├── services/           # Microservices
│   ├── api-gateway/   # API Gateway
│   ├── user-service/  # User service
│   ├── order-service/ # Order service
│   ├── inventory-service/ # Inventory service
│   └── payment-service/   # Payment service
├── pkg/               # Common packages
│   ├── jwt/          # JWT authentication
│   ├── redisclient/  # Redis client with SOLID architecture
│   └── kafka/        # Kafka clients
├── frontend/          # React application
└── proto/            # Protobuf definitions
```

### Adding New Services

1. Create service in `services/` folder
2. Add to `docker-compose.yml`
3. Update API Gateway if necessary
4. Add environment variables

## Monitoring and Logging

- **Zap** - structured logging in Go services
- **Health checks** - service health verification
- **Kafka UI** - Kafka message monitoring

## Production Security

- Change JWT secret keys
- Configure HTTPS
- Add rate limiting
- Configure security monitoring
- Use Kubernetes secrets for sensitive data
