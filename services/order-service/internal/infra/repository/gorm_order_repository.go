package repository

import (
	"context"
	"errors"

	domainerrors "github.com/kubernetestest/ecommerce-platform/services/order-service/internal/domain/errors"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/domain/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormOrderRepository struct {
	db *gorm.DB
}

func NewGormOrderRepository(db *gorm.DB) *GormOrderRepository {
	return &GormOrderRepository{db: db}
}

func (r *GormOrderRepository) Create(ctx context.Context, order *models.Order) error {
	// GORM automatically saves Order + OrderItems with FullSaveAssociations
	result := r.db.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Create(order)
	return result.Error
}

func (r *GormOrderRepository) GetByID(ctx context.Context, id string) (*models.Order, error) {
	var order models.Order
	result := r.db.WithContext(ctx).Preload("Items").First(&order, "id = ?", id)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, domainerrors.ErrOrderNotFound
	}

	return &order, result.Error
}

func (r *GormOrderRepository) GetByUserID(ctx context.Context, userID string, page, limit int) ([]*models.Order, int64, error) {
	var orders []*models.Order
	var total int64

	// Count total records
	if err := r.db.WithContext(ctx).Model(&models.Order{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get orders with pagination
	offset := (page - 1) * limit
	result := r.db.WithContext(ctx).
		Preload("Items").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&orders)

	return orders, total, result.Error
}

func (r *GormOrderRepository) Update(ctx context.Context, order *models.Order) error {
	// Save Order + all related OrderItems
	result := r.db.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Save(order)
	return result.Error
}

func (r *GormOrderRepository) Delete(ctx context.Context, id string) error {
	// GORM automatically deletes related OrderItems with OnDelete:CASCADE
	result := r.db.WithContext(ctx).Delete(&models.Order{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domainerrors.ErrOrderNotFound
	}
	return nil
}

func (r *GormOrderRepository) GetByStatus(ctx context.Context, status models.OrderStatus) ([]*models.Order, error) {
	var orders []*models.Order
	result := r.db.WithContext(ctx).
		Preload("Items").
		Where("status = ?", status).
		Order("created_at DESC").
		Find(&orders)

	return orders, result.Error
}

// AutoMigrate creates tables
func (r *GormOrderRepository) AutoMigrate() error {
	return r.db.AutoMigrate(&models.Order{}, &models.OrderItem{})
}

// NextOrderNumber returns next sequential number per user (transaction-safe)
func (r *GormOrderRepository) NextOrderNumber(ctx context.Context, userID string) (int64, error) {
	var next int64
	// Use advisory lock to serialize allocation per user and avoid aggregate FOR UPDATE
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Lock on user key within transaction scope
		if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", userID).Error; err != nil {
			return err
		}
		// Read last order for user with row-level lock
		var last models.Order
		err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", userID).
			Order("number DESC").
			Limit(1).
			Take(&last).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				next = 1
				return nil
			}
			return err
		}
		next = last.Number + 1
		return nil
	})
	if err != nil {
		return 0, err
	}
	return next, nil
}
