# Configuration Files

This directory contains environment configuration files for all services in the ecommerce platform.

## 📁 Structure

```
configs/
├── services/           # Service-specific configurations
│   ├── *.env.example  # Example files (committed to git)
│   └── *.env          # Actual files (ignored by git)
├── shared/            # Shared configurations
│   ├── *.env.example  # Example files (committed to git)
│   └── *.env          # Actual files (ignored by git)
└── README.md          # This file
```

## 🚀 Quick Setup

### Windows (PowerShell)
```powershell
.\scripts\setup-env.ps1
```

### Linux/macOS (Bash)
```bash
chmod +x scripts/setup-env.sh
./scripts/setup-env.sh
```

## 📋 Manual Setup

If you prefer to set up manually:

1. Copy all `.env.example` files to `.env` files:
   ```bash
   # Service configurations
   cp configs/services/*.env.example configs/services/*.env
   
   # Shared configurations  
   cp configs/shared/*.env.example configs/shared/*.env
   ```

2. Review and adjust values in the created `.env` files

3. Start services:
   ```bash
   docker-compose up
   ```

## 🔧 Configuration Files

### Service-specific files
- **`inventory-service.env`** - Inventory service configuration
- **`user-service.env`** - User service configuration  
- **`order-service.env`** - Order service configuration
- **`payment-service.env`** - Payment service configuration
- **`api-gateway.env`** - API Gateway configuration

### Shared files
- **`database.env`** - Database connection settings
- **`kafka.env`** - Kafka broker settings
- **`redis.env`** - Redis connection settings
- **`monitoring.env`** - Monitoring and metrics settings
- **`logging.env`** - Logging configuration

## 🎯 Benefits

✅ **Clear separation** - Each service sees only its own parameters  
✅ **Security** - No unnecessary environment variables  
✅ **Scalability** - Easy to add new services  
✅ **Debugging** - Easier to find configuration issues  
✅ **CI/CD** - Simpler to deploy individual services  

## 🔒 Security Notes

- `.env` files are ignored by git and should never be committed
- `.env.example` files are safe to commit as they contain no secrets
- Always review configuration values before deployment
- Use strong passwords and secrets in production

## 🐳 Docker Integration

Docker Compose automatically loads the configuration files:

```yaml
services:
  inventory-service:
    env_file:
      - configs/shared/database.env
      - configs/shared/redis.env
      - configs/shared/kafka.env
      - configs/shared/monitoring.env
      - configs/shared/logging.env
      - configs/services/inventory-service.env
```

## 🆘 Troubleshooting

### Missing environment variables
If you see "environment variable not found" errors:
1. Check that all `.env` files exist
2. Verify the variable is defined in the correct file
3. Ensure the variable name matches exactly

### Configuration not loading
1. Verify file permissions
2. Check for syntax errors in `.env` files
3. Ensure Docker Compose is reading the correct files

### Service startup issues
1. Check service-specific logs: `docker-compose logs <service-name>`
2. Verify all dependencies are running
3. Check port conflicts
