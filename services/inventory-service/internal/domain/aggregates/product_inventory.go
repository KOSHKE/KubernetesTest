package aggregates

import (
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

// Validate validates the aggregate by delegating to entity validations
func (p *ProductInventory) Validate() error {
	// Validate product entity
	if p.Product != nil {
		if err := p.Product.Validate(); err != nil {
			return err
		}
	}

	// Validate stock entity (if exists)
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
