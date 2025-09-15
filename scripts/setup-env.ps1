# PowerShell script to setup environment files
Write-Host "Setting up environment files..." -ForegroundColor Green

# Get the script directory and go to project root
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir
Set-Location $ProjectRoot

# Create configs directory if it doesn't exist
if (!(Test-Path "configs/services")) {
    New-Item -ItemType Directory -Path "configs/services" -Force
}
if (!(Test-Path "configs/shared")) {
    New-Item -ItemType Directory -Path "configs/shared" -Force
}

# Copy example files to actual env files
Write-Host "Creating service configuration files..." -ForegroundColor Yellow
Copy-Item "configs/services/inventory-service.env.example" "configs/services/inventory-service.env" -Force
Copy-Item "configs/services/user-service.env.example" "configs/services/user-service.env" -Force
Copy-Item "configs/services/order-service.env.example" "configs/services/order-service.env" -Force
Copy-Item "configs/services/payment-service.env.example" "configs/services/payment-service.env" -Force
Copy-Item "configs/services/api-gateway.env.example" "configs/services/api-gateway.env" -Force

Write-Host "Creating shared configuration files..." -ForegroundColor Yellow
Copy-Item "configs/shared/database.env.example" "configs/shared/database.env" -Force
Copy-Item "configs/shared/kafka.env.example" "configs/shared/kafka.env" -Force
Copy-Item "configs/shared/redis.env.example" "configs/shared/redis.env" -Force
Copy-Item "configs/shared/monitoring.env.example" "configs/shared/monitoring.env" -Force
Copy-Item "configs/shared/logging.env.example" "configs/shared/logging.env" -Force

Write-Host ""
Write-Host "Environment files created successfully!" -ForegroundColor Green
Write-Host ""
Write-Host "Files created:" -ForegroundColor Cyan
Write-Host "Service configurations:" -ForegroundColor White
Write-Host "   - configs/services/inventory-service.env" -ForegroundColor Gray
Write-Host "   - configs/services/user-service.env" -ForegroundColor Gray
Write-Host "   - configs/services/order-service.env" -ForegroundColor Gray
Write-Host "   - configs/services/payment-service.env" -ForegroundColor Gray
Write-Host "   - configs/services/api-gateway.env" -ForegroundColor Gray
Write-Host ""
Write-Host "Shared configurations:" -ForegroundColor White
Write-Host "   - configs/shared/database.env" -ForegroundColor Gray
Write-Host "   - configs/shared/kafka.env" -ForegroundColor Gray
Write-Host "   - configs/shared/redis.env" -ForegroundColor Gray
Write-Host "   - configs/shared/monitoring.env" -ForegroundColor Gray
Write-Host "   - configs/shared/logging.env" -ForegroundColor Gray
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "1. Review and adjust values in the created .env files if needed" -ForegroundColor White
Write-Host "2. Start services with: docker-compose up" -ForegroundColor White
Write-Host ""
Write-Host "Tip: You can edit the .env files to customize your configuration" -ForegroundColor Magenta
