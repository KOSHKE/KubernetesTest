package repository

import (
	"context"

	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/services/order-service/internal/domain/aggregates"
	"ecommerce-platform/services/order-service/internal/domain/entities"
	"ecommerce-platform/services/order-service/internal/domain/ports/repository"
	orderValueObjects "ecommerce-platform/services/order-service/internal/domain/valueobjects"
	migration "ecommerce-platform/services/order-service/internal/infra/migration"

	"gorm.io/gorm"
)

// Use models from migration package
type OrderRecord = migration.OrderRecord
type OrderItemRecord = migration.OrderItemRecord

func recordFromEntity(order *aggregates.Order) (OrderRecord, []OrderItemRecord) {
	orderRec := OrderRecord{
		ID:              order.ID,
		UserID:          order.UserID,
		Status:          string(order.Status),
		ShippingAddress: order.ShippingAddress.Value,
		Currency:        order.Currency.Code,
		TotalAmount:     order.TotalAmount.Amount,
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
	}

	itemRecs := make([]OrderItemRecord, len(order.Items))
	for i, item := range order.Items {
		itemRecs[i] = OrderItemRecord{
			OrderID:     order.ID,
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice.Amount,
			Currency:    item.UnitPrice.Currency.Code,
		}
	}

	return orderRec, itemRecs
}

func entityFromRecord(orderRec OrderRecord, itemRecs []OrderItemRecord) (*aggregates.Order, error) {
	// Data from DB is already validated, so we can safely ignore errors
	currency, _ := valueobjects.NewCurrency(orderRec.Currency)
	totalAmount := valueobjects.NewMoney(orderRec.TotalAmount, currency)
	shippingAddress, _ := orderValueObjects.NewShippingAddress(orderRec.ShippingAddress)
	status := orderValueObjects.OrderStatus(orderRec.Status)

	items := make([]*entities.OrderItem, len(itemRecs))
	for i, itemRec := range itemRecs {
		itemCurrency, _ := valueobjects.NewCurrency(itemRec.Currency)
		unitPrice := valueobjects.NewMoney(itemRec.UnitPrice, itemCurrency)

		items[i] = &entities.OrderItem{
			ProductID:   itemRec.ProductID,
			ProductName: itemRec.ProductName,
			Quantity:    itemRec.Quantity,
			UnitPrice:   unitPrice,
		}
	}

	return &aggregates.Order{
		ID:              orderRec.ID,
		UserID:          orderRec.UserID,
		Status:          status,
		Items:           items,
		ShippingAddress: shippingAddress,
		Currency:        currency,
		TotalAmount:     totalAmount,
		CreatedAt:       orderRec.CreatedAt,
		UpdatedAt:       orderRec.UpdatedAt,
	}, nil
}

// convertOrderRecords converts slice of OrderRecord to domain objects
func convertOrderRecords(orderRecs []OrderRecord) ([]*aggregates.Order, error) {
	orders := make([]*aggregates.Order, len(orderRecs))
	for i, orderRec := range orderRecs {
		orders[i], _ = entityFromRecord(orderRec, orderRec.Items)
	}
	return orders, nil
}

// GormOrderRepository implements the OrderRepository interface using GORM
type GormOrderRepository struct {
	db *gorm.DB
}

// NewOrderRepository creates a new order repository
func NewOrderRepository(db *gorm.DB) repository.OrderRepository {
	return &GormOrderRepository{db: db}
}

func (r *GormOrderRepository) Create(ctx context.Context, order *aggregates.Order) error {
	orderRec, itemRecs := recordFromEntity(order)

	// Create order with items using GORM associations
	if err := r.db.WithContext(ctx).Create(&orderRec).Error; err != nil {
		return err
	}

	// Add items to the order using association
	if len(itemRecs) > 0 {
		if err := r.db.WithContext(ctx).Model(&orderRec).Association("Items").Append(itemRecs); err != nil {
			return err
		}
	}

	return nil
}

func (r *GormOrderRepository) GetByID(ctx context.Context, id string) (*aggregates.Order, error) {
	var orderRec OrderRecord
	result := r.db.WithContext(ctx).Preload("Items").First(&orderRec, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}

	order, _ := entityFromRecord(orderRec, orderRec.Items)
	return order, nil
}

func (r *GormOrderRepository) GetByUserID(ctx context.Context, userID string, page, limit int) ([]*aggregates.Order, int64, error) {
	var total int64
	result := r.db.WithContext(ctx).Model(&OrderRecord{}).Where("user_id = ?", userID).Count(&total)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	offset := (page - 1) * limit

	// Fetch orders with items using Preload
	var orderRecs []OrderRecord
	result = r.db.WithContext(ctx).Preload("Items").Where("user_id = ?", userID).Order("created_at DESC").Offset(offset).Limit(limit).Find(&orderRecs)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	// Convert to domain objects
	orders, err := convertOrderRecords(orderRecs)
	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *GormOrderRepository) Update(ctx context.Context, order *aggregates.Order) error {
	orderRec, itemRecs := recordFromEntity(order)

	// Start transaction for atomic update
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update order record
		if err := tx.Save(&orderRec).Error; err != nil {
			return err
		}

		// Replace items using GORM association
		// This will automatically handle DELETE + INSERT efficiently
		if err := tx.Model(&orderRec).Association("Items").Replace(itemRecs); err != nil {
			return err
		}

		return nil
	})
}
