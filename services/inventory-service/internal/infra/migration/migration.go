package migration

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// Migration represents a database migration
type Migration struct {
	Version     int64
	Description string
	Up          func(*gorm.DB) error
	Down        func(*gorm.DB) error
}

// MigrationService handles database migrations
type MigrationService struct {
	db *gorm.DB
}

// NewMigrationService creates a new migration service
func NewMigrationService(db *gorm.DB) *MigrationService {
	return &MigrationService{db: db}
}

// Migrate runs all pending migrations
func (s *MigrationService) Migrate(ctx context.Context) error {
	// Create migrations table if not exists
	if err := s.createMigrationsTable(); err != nil {
		return err
	}

	// Get all migrations
	migrations := s.getMigrations()

	// Get applied migrations
	applied, err := s.getAppliedMigrations()
	if err != nil {
		return err
	}

	// Apply pending migrations
	for _, migration := range migrations {
		if !applied[migration.Version] {
			if err := s.applyMigration(migration); err != nil {
				return err
			}
		}
	}

	return nil
}

// createMigrationsTable creates the migrations tracking table
func (s *MigrationService) createMigrationsTable() error {
	return s.db.AutoMigrate(&MigrationRecord{})
}

// getAppliedMigrations returns map of applied migration versions
func (s *MigrationService) getAppliedMigrations() (map[int64]bool, error) {
	var records []MigrationRecord
	if err := s.db.Find(&records).Error; err != nil {
		return nil, err
	}

	applied := make(map[int64]bool)
	for _, record := range records {
		applied[record.Version] = true
	}
	return applied, nil
}

// applyMigration applies a single migration
func (s *MigrationService) applyMigration(migration Migration) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := migration.Up(tx); err != nil {
			return err
		}

		record := MigrationRecord{
			Version:     migration.Version,
			Description: migration.Description,
		}
		return tx.Create(&record).Error
	})
}

// getMigrations returns all available migrations
func (s *MigrationService) getMigrations() []Migration {
	return []Migration{
		{
			Version:     1,
			Description: "Create products table",
			Up: func(db *gorm.DB) error {
				return db.AutoMigrate(&ProductRecord{})
			},
			Down: func(db *gorm.DB) error {
				return db.Migrator().DropTable(&ProductRecord{})
			},
		},
		{
			Version:     2,
			Description: "Create stocks table",
			Up: func(db *gorm.DB) error {
				return db.AutoMigrate(&StockRecord{})
			},
			Down: func(db *gorm.DB) error {
				return db.Migrator().DropTable(&StockRecord{})
			},
		},
		{
			Version:     3,
			Description: "Create outbox_events table",
			Up: func(db *gorm.DB) error {
				return db.AutoMigrate(&OutboxRecord{})
			},
			Down: func(db *gorm.DB) error {
				return db.Migrator().DropTable(&OutboxRecord{})
			},
		},
	}
}

// SeedData seeds the database with initial data
func (s *MigrationService) SeedData(ctx context.Context) error {
	// Seed products
	products := []*ProductRecord{
		{
			ID:       "prod-1",
			Name:     "Wireless Headphones",
			Price:    9999,
			Currency: "USD",
			ImageURL: "/images/headphones.jpg",
		},
		{
			ID:       "prod-2",
			Name:     "Smart Watch",
			Price:    19999,
			Currency: "USD",
			ImageURL: "/images/smartwatch.jpg",
		},
		{
			ID:       "prod-3",
			Name:     "Coffee Mug",
			Price:    1599,
			Currency: "USD",
			ImageURL: "/images/mug.jpg",
		},
	}

	for _, product := range products {
		if err := s.db.WithContext(ctx).FirstOrCreate(product, "id = ?", product.ID).Error; err != nil {
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
		if err := s.db.WithContext(ctx).FirstOrCreate(stock, "product_id = ?", stock.ProductID).Error; err != nil {
			return fmt.Errorf("failed to seed stock for product %s: %w", stock.ProductID, err)
		}
	}

	return nil
}
