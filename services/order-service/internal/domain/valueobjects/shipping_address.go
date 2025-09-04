package valueobjects

import (
	"strings"
)

// ShippingAddress represents a shipping address for orders
type ShippingAddress struct {
	Value string
}

// NewShippingAddress creates a new ShippingAddress instance
func NewShippingAddress(address string) (ShippingAddress, error) {
	return ShippingAddress{Value: strings.TrimSpace(address)}, nil
}

// String returns the string representation of the address
func (sa ShippingAddress) String() string {
	return sa.Value
}

// IsEmpty checks if the shipping address is empty
func (sa ShippingAddress) IsEmpty() bool {
	return sa.Value == ""
}
