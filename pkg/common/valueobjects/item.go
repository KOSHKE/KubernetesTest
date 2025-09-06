package valueobjects

import "ecommerce-platform/pkg/common/errors"

// Item represents a basic item with product ID and quantity
type Item struct {
	ProductID string
	Quantity  int32
}

// NewItem creates a new Item value object
func NewItem(productID string, quantity int32) (*Item, error) {
	item := &Item{
		ProductID: productID,
		Quantity:  quantity,
	}

	if err := item.Validate(); err != nil {
		return nil, err
	}

	return item, nil
}

// Validate validates item data
func (i *Item) Validate() error {
	if i.ProductID == "" {
		return errors.ErrInvalidProductID
	}
	if i.Quantity <= 0 {
		return errors.ErrInvalidQuantity
	}
	return nil
}

// Equals checks if two Items are equal
func (i *Item) Equals(other *Item) bool {
	if other == nil {
		return false
	}
	return i.ProductID == other.ProductID && i.Quantity == other.Quantity
}
