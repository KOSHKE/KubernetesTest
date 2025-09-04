package migration

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// MigrationService handles database migrations
type MigrationService struct {
	db *gorm.DB
}

// NewMigrationService creates a new migration service
func NewMigrationService(db *gorm.DB) *MigrationService {
	return &MigrationService{db: db}
}

// RunMigrations runs all database migrations
func (m *MigrationService) RunMigrations(ctx context.Context) error {
	// Auto migrate all tables
	if err := m.db.WithContext(ctx).AutoMigrate(
		&ProductRecord{},
		&StockRecord{},
	); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// SeedData seeds the database with initial data
func (m *MigrationService) SeedData(ctx context.Context) error {
	// Seed products
	products := []*ProductRecord{
		{
			ID:          "prod-1",
			Name:        "Wireless Headphones",
			Description: "High-quality wireless headphones with noise cancellation",
			PriceMinor:  9999,
			Currency:    "USD",
			ImageURL:    "/images/headphones.jpg",
			IsActive:    true,
		},
		{
			ID:          "prod-2",
			Name:        "Smart Watch",
			Description: "Fitness tracking smart watch with heart rate monitor",
			PriceMinor:  19999,
			Currency:    "USD",
			ImageURL:    "/images/smartwatch.jpg",
			IsActive:    true,
		},
		{
			ID:          "prod-3",
			Name:        "Coffee Mug",
			Description: "Ceramic coffee mug with custom design",
			PriceMinor:  1599,
			Currency:    "USD",
			ImageURL:    "/images/mug.jpg",
			IsActive:    true,
		},
	}

	for _, product := range products {
		if err := m.db.WithContext(ctx).FirstOrCreate(product, "id = ?", product.ID).Error; err != nil {
			return fmt.Errorf("failed to seed product %s: %w", product.ID, err)
		}
	}

	// Seed stocks
	stocks := []*StockRecord{
		{
			ProductID:         "prod-1",
			AvailableQuantity: 3,
			ReservedQuantity:  0,
		},
		{
			ProductID:         "prod-2",
			AvailableQuantity: 1,
			ReservedQuantity:  0,
		},
		{
			ProductID:         "prod-3",
			AvailableQuantity: 2,
			ReservedQuantity:  0,
		},
	}

	for _, stock := range stocks {
		if err := m.db.WithContext(ctx).FirstOrCreate(stock, "product_id = ?", stock.ProductID).Error; err != nil {
			return fmt.Errorf("failed to seed stock for product %s: %w", stock.ProductID, err)
		}
	}

	return nil
}
