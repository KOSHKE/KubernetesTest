# Kubernetes Test Project

A microservices-based order management system built with Go, React, and modern cloud-native technologies.

## 🚀 Quick Start

```bash
# Clone the repository
git clone <repository-url>
cd KubernetesTest

# Create environment file
cp .env.example .env
# Edit .env with your configuration

# Start all services
docker-compose up -d

# Access the application
# Frontend: http://localhost:3001
# API Gateway: http://localhost:8080
```

## ✨ Features

- **JWT Authentication** with access/refresh tokens
- **Microservices Architecture** with Go services
- **Real-time Communication** via Kafka
- **Modern Frontend** with React + TypeScript
- **Containerized** with Docker
- **Database** with PostgreSQL
- **Caching** with Redis
- **Message Queue** with Kafka

## 🏗️ Architecture

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
```

### Services

- **API Gateway** - HTTP API with authentication middleware
- **User Service** - User management and JWT authentication
- **Order Service** - Order processing and management
- **Inventory Service** - Product catalog and stock management
- **Payment Service** - Payment processing and refunds

## 🔐 JWT Authentication

The system implements secure JWT-based authentication:

- **Access Tokens** - Short-lived (15 minutes) for API requests
- **Refresh Tokens** - Long-lived (7 days) stored in Redis
- **Automatic Refresh** - Transparent token renewal on frontend
- **Secure Logout** - Token revocation capability

### Authentication Flow

1. User logs in → receives access + refresh tokens
2. Access token used for API requests
3. When expired → automatic refresh using refresh token
4. On logout → refresh token revoked from Redis

## 📡 API Endpoints

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

## 🛠️ Development

### Prerequisites

- Docker & Docker Compose
- Go 1.21+
- Node.js 18+
- Make (optional)

### Environment Variables

Create a `.env` file with the following variables:

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=testdb
DB_USER=admin
DB_PASSWORD=password

# JWT
JWT_SECRET=your-super-secret-jwt-key-change-in-production
JWT_REFRESH_SECRET=your-super-secret-refresh-key-change-in-production
ACCESS_TOKEN_TTL=15m
REFRESH_TOKEN_TTL=168h

# Redis
REDIS_URL=redis:6379

# Kafka
KAFKA_BROKERS=kafka:9092
KAFKA_ADVERTISED_LISTENERS=PLAINTEXT://localhost:9092
KAFKA_KRAFT_CLUSTER_ID=your-cluster-id

# Services
USER_SERVICE_HTTP_PORT=50051
VITE_API_URL=http://localhost:8080/api/v1
```

### Commands

```bash
# Start development environment
make dev

# Start specific service
make dev-service SERVICE=user-service

# View logs
make logs SERVICE=api-gateway

# Stop all services
make down

# Clean up
make clean

# Run tests
make test

# Build production images
make build
```

### Service Development

Each service supports hot reload with Air:

```bash
# Navigate to service directory
cd services/user-service

# Start with hot reload
air
```

## 📊 Monitoring

- **Health Checks** - Service health endpoints
- **Structured Logging** - Zap logger with structured fields
- **Kafka UI** - Message monitoring at http://localhost:8085
- **Database** - Direct access to PostgreSQL and Redis

## 🔒 Security Features

- JWT token validation on all protected routes
- Refresh token storage in Redis with revocation
- CORS configuration for frontend origins
- Input validation and sanitization
- Secure password hashing
- Rate limiting ready (can be added)

## 🚀 Production Deployment

### Security Checklist

- [ ] Change default JWT secrets
- [ ] Configure HTTPS/TLS
- [ ] Set up proper CORS origins
- [ ] Add rate limiting
- [ ] Configure monitoring and alerting
- [ ] Use Kubernetes secrets for sensitive data
- [ ] Set up proper logging aggregation

### Kubernetes Ready

The project structure is designed for easy Kubernetes deployment:

- Containerized services
- Health check endpoints
- Environment-based configuration
- Stateless service design

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

For questions and support:

- Check the documentation in `.cursor/project.md`
- Review the service logs: `docker-compose logs -f [service-name]`
- Check service health: `curl http://localhost:8080/health`
