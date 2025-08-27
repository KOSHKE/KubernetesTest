package repository

import (
	"errors"
	"fmt"
	"time"

	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/domain/aggregates"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/domain/entities"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/domain/valueobjects"
)

// GORMOrder represents the Order model for GORM
type GORMOrder struct {
	ID              string                   `gorm:"primaryKey;type:varchar(255)"`
	UserID          string                   `gorm:"type:varchar(255);not null;uniqueIndex:idx_user_number"`
	Number          int64                    `gorm:"type:bigint;not null;uniqueIndex:idx_user_number"`
	Status          valueobjects.OrderStatus `gorm:"type:varchar(50);not null;default:'PENDING'"`
	ShippingAddress string                   `gorm:"type:text;not null"`
	Currency        string                   `gorm:"type:varchar(3);not null;default:'USD'"`
	TotalAmount     int64                    `gorm:"type:bigint;not null;default:0"` // Store amount in minor units
	CreatedAt       time.Time                `gorm:"autoCreateTime"`
	UpdatedAt       time.Time                `gorm:"autoUpdateTime"`
}

// TableName specifies the table name for GORM
func (GORMOrder) TableName() string {
	return "orders"
}

// GORMOrderItem represents the OrderItem model for GORM
type GORMOrderItem struct {
	ID          uint   `gorm:"primaryKey;autoIncrement"`
	OrderID     string `gorm:"type:varchar(255);not null;uniqueIndex:idx_order_product"`
	ProductID   string `gorm:"type:varchar(255);not null;uniqueIndex:idx_order_product"`
	ProductName string `gorm:"type:varchar(500);not null"`
	Quantity    int32  `gorm:"type:int;not null"`
	UnitPrice   int64  `gorm:"type:bigint;not null"` // Amount in minor units
	Currency    string `gorm:"type:varchar(3);not null"`
}

// Validate ensures OrderID is set before creation
func (oi *GORMOrderItem) Validate() error {
	if oi.OrderID == "" {
		return errors.New("OrderID is required")
	}
	return nil
}

// TableName specifies the table name for GORM
func (GORMOrderItem) TableName() string {
	return "order_items"
}

// ToDomain converts GORMOrder to domain Order aggregate
func (g *GORMOrder) ToDomain() (*aggregates.Order, error) {
	items := make([]*entities.OrderItem, 0) // Initialize as empty slice

	// Create Money value object from stored amount and currency
	totalAmount, err := valueobjects.NewMoney(g.TotalAmount, g.Currency)
	if err != nil {
		return nil, fmt.Errorf("failed to create total amount money: %w", err)
	}

	return &aggregates.Order{
		ID:              g.ID,
		UserID:          g.UserID,
		Number:          g.Number,
		Status:          g.Status,
		Items:           items,
		ShippingAddress: g.ShippingAddress,
		Currency:        g.Currency,
		TotalAmount:     totalAmount,
		CreatedAt:       g.CreatedAt,
		UpdatedAt:       g.UpdatedAt,
	}, nil
}

// FromDomain converts domain Order aggregate to GORMOrder
func (g *GORMOrder) FromDomain(order *aggregates.Order) {
	g.ID = order.ID
	g.UserID = order.UserID
	g.Number = order.Number
	g.Status = order.Status
	g.ShippingAddress = order.ShippingAddress
	g.Currency = order.Currency
	g.TotalAmount = order.TotalAmount.Amount // Extract amount from Money value object
	g.CreatedAt = order.CreatedAt
	g.UpdatedAt = order.UpdatedAt
	// Items will be handled separately by repository
}
