package aggregates

import (
	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/services/inventory-service/internal/domain/entities"
	stockvalueobjects "ecommerce-platform/services/inventory-service/internal/domain/valueobjects"
)

// Product represents the product aggregate root
type Product struct {
	*entities.Product
	Stock *stockvalueobjects.Stock
}

// NewProduct creates a new product aggregate
func NewProduct(
	id, name, description string,
	price *valueobjects.Money,
	imageURL string,
	availableQuantity, reservedQuantity int32,
) *Product {
	return &Product{
		Product: entities.NewProduct(id, name, description, price, imageURL, true),
		Stock:   stockvalueobjects.NewStock(availableQuantity, reservedQuantity),
	}
}

// ReserveStock reserves the specified quantity of stock
func (p *Product) ReserveStock(quantity int32) error {
	if !p.Stock.CanReserve(quantity) {
		return errors.ErrInsufficientStock
	}

	p.Stock = p.Stock.Reserve(quantity)
	return nil
}

// ReleaseStock releases the specified quantity of stock
func (p *Product) ReleaseStock(quantity int32) {
	p.Stock = p.Stock.Release(quantity)
}

// CommitStock commits the reserved quantity (removes it completely)
func (p *Product) CommitStock(quantity int32) {
	p.Stock = p.Stock.Commit(quantity)
}

// AddStock adds quantity to available stock
func (p *Product) AddStock(quantity int32) {
	p.Stock = p.Stock.AddAvailable(quantity)
}

// GetAvailableQuantity returns the available quantity
func (p *Product) GetAvailableQuantity() int32 {
	return p.Stock.AvailableQuantity
}

// GetReservedQuantity returns the reserved quantity
func (p *Product) GetReservedQuantity() int32 {
	return p.Stock.ReservedQuantity
}

// GetTotalQuantity returns the total quantity
func (p *Product) GetTotalQuantity() int32 {
	return p.Stock.TotalQuantity()
}

// CanReserve checks if the requested quantity can be reserved
func (p *Product) CanReserve(quantity int32) bool {
	return p.Stock.CanReserve(quantity)
}

// IsInStock checks if product has available stock
func (p *Product) IsInStock() bool {
	return p.Stock.AvailableQuantity > 0
}

// IsAvailableForPurchase checks if product is available for purchase
func (p *Product) IsAvailableForPurchase() bool {
	return p.IsActive && p.IsInStock()
}

// UpdateStock updates the stock information
func (p *Product) UpdateStock(availableQuantity, reservedQuantity int32) {
	p.Stock = stockvalueobjects.NewStock(availableQuantity, reservedQuantity)
}

// Business rules validation
func (p *Product) Validate() error {
	if p.ID == "" {
		return errors.ErrInvalidProductID
	}
	if p.Name == "" {
		return errors.ErrInvalidProductName
	}
	if p.Price == nil {
		return errors.ErrInvalidProductPrice
	}
	if p.Stock == nil {
		return errors.ErrInvalidStock
	}
	return nil
}
