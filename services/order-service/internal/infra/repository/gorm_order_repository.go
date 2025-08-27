package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/domain/aggregates"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/domain/entities"
	domainerrors "github.com/kubernetestest/ecommerce-platform/services/order-service/internal/domain/errors"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/domain/valueobjects"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormOrderRepository struct {
	db *gorm.DB
}

func NewGormOrderRepository(db *gorm.DB) *GormOrderRepository {
	return &GormOrderRepository{db: db}
}

func (r *GormOrderRepository) Create(ctx context.Context, order *aggregates.Order) error {
	gormOrder := &GORMOrder{}
	gormOrder.FromDomain(order)

	// Create order first
	err := r.db.WithContext(ctx).Create(gormOrder).Error
	if err != nil {
		return err
	}

	// Create order items separately
	for _, item := range order.Items {
		gormItem := GORMOrderItem{
			OrderID:     order.ID,
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice.Amount,
			Currency:    item.UnitPrice.Currency,
		}

		err := r.db.WithContext(ctx).Create(&gormItem).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *GormOrderRepository) GetByID(ctx context.Context, id string) (*aggregates.Order, error) {
	var gormOrder GORMOrder
	result := r.db.WithContext(ctx).First(&gormOrder, "id = ?", id)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, domainerrors.ErrOrderNotFound
	}

	if result.Error != nil {
		return nil, result.Error
	}

	order, err := gormOrder.ToDomain()
	if err != nil {
		return nil, fmt.Errorf("failed to convert GORM order to domain: %w", err)
	}

	// Load order items separately
	items, err := r.loadOrderItems(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load order items: %w", err)
	}
	order.Items = items

	return order, nil
}

// loadOrderItems loads order items for a specific order
func (r *GormOrderRepository) loadOrderItems(ctx context.Context, orderID string) ([]*entities.OrderItem, error) {
	var gormItems []GORMOrderItem
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Find(&gormItems).Error
	if err != nil {
		return nil, err
	}

	items := make([]*entities.OrderItem, len(gormItems))
	for i, gormItem := range gormItems {
		unitPrice, err := valueobjects.NewMoney(gormItem.UnitPrice, gormItem.Currency)
		if err != nil {
			return nil, fmt.Errorf("failed to create unit price money for item %s: %w", gormItem.ProductID, err)
		}

		items[i] = &entities.OrderItem{
			ProductID:   gormItem.ProductID,
			ProductName: gormItem.ProductName,
			Quantity:    gormItem.Quantity,
			UnitPrice:   unitPrice,
		}
	}

	return items, nil
}

func (r *GormOrderRepository) GetByUserID(ctx context.Context, userID string, page, limit int) ([]*aggregates.Order, int64, error) {
	var gormOrders []GORMOrder
	var total int64

	if err := r.db.WithContext(ctx).Model(&GORMOrder{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&gormOrders)

	orders := make([]*aggregates.Order, len(gormOrders))
	for i, gormOrder := range gormOrders {
		order, err := gormOrder.ToDomain()
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert GORM order to domain: %w", err)
		}

		// Load order items separately
		items, err := r.loadOrderItems(ctx, order.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to load order items for order %s: %w", order.ID, err)
		}
		order.Items = items

		orders[i] = order
	}

	return orders, total, result.Error
}

func (r *GormOrderRepository) Update(ctx context.Context, order *aggregates.Order) error {
	gormOrder := &GORMOrder{}
	gormOrder.FromDomain(order)

	// Update order
	err := r.db.WithContext(ctx).Save(gormOrder).Error
	if err != nil {
		return err
	}

	// Delete existing items and recreate them
	err = r.db.WithContext(ctx).Where("order_id = ?", order.ID).Delete(&GORMOrderItem{}).Error
	if err != nil {
		return err
	}

	// Create new items
	for _, item := range order.Items {
		gormItem := GORMOrderItem{
			OrderID:     order.ID,
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice.Amount,
			Currency:    item.UnitPrice.Currency,
		}

		err := r.db.WithContext(ctx).Create(&gormItem).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *GormOrderRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&GORMOrder{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domainerrors.ErrOrderNotFound
	}
	return nil
}

// AutoMigrate creates tables
func (r *GormOrderRepository) AutoMigrate() error {
	return r.db.AutoMigrate(&GORMOrder{}, &GORMOrderItem{})
}

// NextOrderNumber returns next sequential number per user (transaction-safe)
func (r *GormOrderRepository) NextOrderNumber(ctx context.Context, userID string) (int64, error) {
	var next int64
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var last GORMOrder
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
