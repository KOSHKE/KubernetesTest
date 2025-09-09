package aggregates

import (
	"errors"

	"ecommerce-platform/services/inventory-service/internal/domain/entities"
)

// ProductInventory represents an aggregate of Product and Stock entities
// This aggregate solves the N+1 problem by combining product and stock data
// Uses composition for explicit control over fields and methods
type ProductInventory struct {
	Product *entities.Product
	Stock   *entities.Stock
}

// NewProductInventory creates a new ProductInventory aggregate
func NewProductInventory(
	product *entities.Product,
	stock *entities.Stock,
) *ProductInventory {
	return &ProductInventory{
		Product: product,
		Stock:   stock,
	}
}

// GetTotalQuantity returns the total quantity (available + reserved)
func (p *ProductInventory) GetTotalQuantity() int32 {
	if p.Stock == nil {
		return 0
	}
	return p.Stock.AvailableQuantity + p.Stock.ReservedQuantity
}

// HasStock checks if the product has any stock
func (p *ProductInventory) HasStock() bool {
	return p.GetTotalQuantity() > 0
}

// CanReserve checks if the requested quantity can be reserved
func (p *ProductInventory) CanReserve(quantity int32) bool {
	if p.Stock == nil {
		return false
	}
	return p.Stock.AvailableQuantity >= quantity
}

// Validate validates the aggregate data
func (p *ProductInventory) Validate() error {
	// Validate product fields
	if p.Product == nil {
		return errors.New("product cannot be nil")
	}
	if err := p.Product.Validate(); err != nil {
		return err
	}

	// Validate stock fields (if exists)
	if p.Stock != nil {
		if err := p.Stock.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// ToProduct converts the aggregate to a Product entity
func (p *ProductInventory) ToProduct() *entities.Product {
	return p.Product
}

// ToStock converts the aggregate to a Stock entity
func (p *ProductInventory) ToStock() *entities.Stock {
	return p.Stock
}
