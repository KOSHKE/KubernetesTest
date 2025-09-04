package migration

import (
	"context"

	"gorm.io/gorm"
)

// Migration represents a database migration
type Migration struct {
	Version     int64
	Description string
	Up          func(*gorm.DB) error
	Down        func(*gorm.DB) error
}

// MigrationService handles database schema migrations
type MigrationService struct {
	db *gorm.DB
}

// NewMigrationService creates new migration service
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
	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Run migration
	if err := migration.Up(tx); err != nil {
		tx.Rollback()
		return err
	}

	// Record migration
	record := MigrationRecord{
		Version:     migration.Version,
		Description: migration.Description,
	}
	if err := tx.Create(&record).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// getMigrations returns all available migrations
func (s *MigrationService) getMigrations() []Migration {
	return []Migration{
		{
			Version:     1,
			Description: "Create users table",
			Up: func(db *gorm.DB) error {
				return db.AutoMigrate(&UserRecord{})
			},
			Down: func(db *gorm.DB) error {
				return db.Migrator().DropTable(&UserRecord{})
			},
		},
	}
}
