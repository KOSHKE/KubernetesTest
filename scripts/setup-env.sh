#!/bin/bash

echo "Setting up environment files..."

# Get the script directory and go to project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT"

# Create configs directory if it doesn't exist
mkdir -p configs/services
mkdir -p configs/shared

# Copy example files to actual env files
echo "Creating service configuration files..."
cp configs/services/inventory-service.env.example configs/services/inventory-service.env
cp configs/services/user-service.env.example configs/services/user-service.env
cp configs/services/order-service.env.example configs/services/order-service.env
cp configs/services/payment-service.env.example configs/services/payment-service.env
cp configs/services/api-gateway.env.example configs/services/api-gateway.env

echo "Creating shared configuration files..."
cp configs/shared/database.env.example configs/shared/database.env
cp configs/shared/kafka.env.example configs/shared/kafka.env
cp configs/shared/redis.env.example configs/shared/redis.env
cp configs/shared/monitoring.env.example configs/shared/monitoring.env
cp configs/shared/logging.env.example configs/shared/logging.env

echo ""
echo "✅ Environment files created successfully!"
echo ""
echo "Files created:"
echo "📁 Service configurations:"
echo "   - configs/services/inventory-service.env"
echo "   - configs/services/user-service.env"
echo "   - configs/services/order-service.env"
echo "   - configs/services/payment-service.env"
echo "   - configs/services/api-gateway.env"
echo ""
echo "📁 Shared configurations:"
echo "   - configs/shared/database.env"
echo "   - configs/shared/kafka.env"
echo "   - configs/shared/redis.env"
echo "   - configs/shared/monitoring.env"
echo "   - configs/shared/logging.env"
echo ""
echo "🔧 Next steps:"
echo "1. Review and adjust values in the created .env files if needed"
echo "2. Start services with: docker-compose up"
echo ""
echo "💡 Tip: You can edit the .env files to customize your configuration"
