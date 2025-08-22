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
- **Redis** - Cache and refresh token storage
- **Kafka** - Asynchronous communication between services
- **Frontend** - React application with TypeScript

## JWT Authentication

### Implemented Components

#### 1. Authentication Libraries

- **`libs/jwt/`** - JWT token management (generation, validation, refresh)
- **`libs/redis/`** - Storage and management of refresh tokens in Redis

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
├── libs/              # Common libraries
│   ├── jwt/          # JWT authentication
│   ├── redis/        # Redis client
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
