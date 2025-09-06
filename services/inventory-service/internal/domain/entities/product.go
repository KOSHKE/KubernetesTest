package entities

import (
	"time"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/common/valueobjects"
)

// Product represents a product entity in the inventory domain
type Product struct {
	ID        string
	Name      string
	Price     valueobjects.Money
	ImageURL  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewProduct creates a new product entity
func NewProduct(
	id, name string,
	price valueobjects.Money,
	imageURL string,
) *Product {
	return &Product{
		ID:        id,
		Name:      name,
		Price:     price,
		ImageURL:  imageURL,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// IsAvailable checks if product is available for purchase
func (p *Product) IsAvailable() bool {
	return p.ID != "" && p.Name != ""
}

// UpdateDetails updates product details
func (p *Product) UpdateDetails(name string, price valueobjects.Money) {
	p.Name = name
	p.Price = price
	p.UpdatedAt = time.Now()
}

// UpdateImageURL updates product image URL
func (p *Product) UpdateImageURL(imageURL string) {
	p.ImageURL = imageURL
	p.UpdatedAt = time.Now()
}

// Validate validates product data
func (p *Product) Validate() error {
	if p.ID == "" {
		return errors.ErrInvalidProductID
	}
	if p.Name == "" {
		return errors.ErrInvalidProductName
	}
	return nil
}
