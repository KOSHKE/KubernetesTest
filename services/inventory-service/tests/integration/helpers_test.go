package integration_test

import (
	"context"
	"testing"

	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/services/inventory-service/internal/application/dto"
	"ecommerce-platform/services/inventory-service/internal/application/services"
	"ecommerce-platform/services/inventory-service/internal/infra/migration"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupTestDB creates a test DB with transaction that will be rolled back
// This ensures complete isolation between tests
func setupTestDB(t *testing.T) (*gorm.DB, func()) {
	// Use PostgreSQL from Docker Compose for realistic integration testing
	// This assumes the database is running via docker-compose up
	dsn := "host=localhost port=5432 user=postgres password=postgres dbname=inventory_test sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	// Test connection
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Ping())

	// Run migrations
	err = db.AutoMigrate(
		&migration.ProductRecord{},
		&migration.StockRecord{},
		&migration.OutboxRecord{},
	)
	require.NoError(t, err)

	// Start transaction
	tx := db.Begin()
	require.NoError(t, tx.Error)

	// Return transaction and cleanup function
	cleanup := func() {
		// Rollback transaction (this undoes all changes)
		tx.Rollback()

		// Close main connection
		sqlDB, err := db.DB()
		if err != nil {
			return
		}
		sqlDB.Close()
	}

	return tx, cleanup
}

// createTestProduct creates a test product with stock using the application service
func createTestProduct(t *testing.T, service *services.InventoryApplicationService, name string, priceAmount int64, stock int32) *dto.ProductResponse {
	ctx := context.Background()
	currency, err := valueobjects.NewCurrency("USD")
	require.NoError(t, err)
	price := valueobjects.NewMoney(priceAmount, currency)

	req := &dto.CreateProductRequest{
		Name:     name,
		Price:    price,
		ImageURL: "https://example.com/image.jpg",
		Stock:    stock,
	}

	product, err := service.CreateProduct(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, product)

	return product
}

// convertToDTOItems converts domain value objects to DTO items
func convertToDTOItems(items []valueobjects.Item) []dto.StockReservationItem {
	dtoItems := make([]dto.StockReservationItem, len(items))
	for i, item := range items {
		dtoItems[i] = dto.StockReservationItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}
	return dtoItems
}
