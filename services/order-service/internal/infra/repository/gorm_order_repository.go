package repository

import (
	"context"
	"errors"

	domainerrors "order-service/internal/domain/errors"
	"order-service/internal/domain/models"

	"gorm.io/gorm"
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
	r.db.WithContext(ctx).Model(&models.Order{}).Where("user_id = ?", userID).Count(&total)

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
	return result.Error
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
