package entities

import (
	"time"

	"ecommerce-platform/pkg/common/valueobjects"
)

// Product represents a product entity in the inventory domain
type Product struct {
	ID          string
	Name        string
	Description string
	Price       *valueobjects.Money
	ImageURL    string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewProduct creates a new product entity
func NewProduct(
	id, name, description string,
	price *valueobjects.Money,
	imageURL string,
	isActive bool,
) *Product {
	return &Product{
		ID:          id,
		Name:        name,
		Description: description,
		Price:       price,
		ImageURL:    imageURL,
		IsActive:    isActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// UpdateDetails updates product details
func (p *Product) UpdateDetails(name, description string, price *valueobjects.Money) {
	p.Name = name
	p.Description = description
	p.Price = price
	p.UpdatedAt = time.Now()
}

// UpdateImageURL updates product image URL
func (p *Product) UpdateImageURL(imageURL string) {
	p.ImageURL = imageURL
	p.UpdatedAt = time.Now()
}

// Activate activates the product
func (p *Product) Activate() {
	p.IsActive = true
	p.UpdatedAt = time.Now()
}

// Deactivate deactivates the product
func (p *Product) Deactivate() {
	p.IsActive = false
	p.UpdatedAt = time.Now()
}

// IsAvailable checks if product is available for purchase
func (p *Product) IsAvailable() bool {
	return p.IsActive
}
