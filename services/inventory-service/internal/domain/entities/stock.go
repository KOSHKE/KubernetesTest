package entities

import (
	"time"

	"ecommerce-platform/pkg/common/errors"
)

// Stock represents a stock entity in the inventory domain
type Stock struct {
	ProductID         string
	AvailableQuantity int32
	ReservedQuantity  int32
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// NewStock creates a new stock entity
func NewStock(productID string, availableQuantity, reservedQuantity int32) *Stock {
	return &Stock{
		ProductID:         productID,
		AvailableQuantity: availableQuantity,
		ReservedQuantity:  reservedQuantity,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
}

// CanReserve checks if the requested quantity can be reserved
func (s *Stock) CanReserve(quantity int32) bool {
	return s.AvailableQuantity >= quantity
}

// CanRelease checks if the requested quantity can be released
func (s *Stock) CanRelease(quantity int32) bool {
	return s.ReservedQuantity >= quantity
}

// CanCommit checks if the requested quantity can be committed
func (s *Stock) CanCommit(quantity int32) bool {
	return s.ReservedQuantity >= quantity
}

// Reserve reserves the specified quantity
func (s *Stock) Reserve(quantity int32) error {
	if !s.CanReserve(quantity) {
		return errors.ErrInsufficientStock
	}

	s.AvailableQuantity -= quantity
	s.ReservedQuantity += quantity
	return nil
}

// Release releases the specified quantity from reserved back to available
func (s *Stock) Release(quantity int32) error {
	if !s.CanRelease(quantity) {
		return errors.ErrInsufficientReservedStock
	}

	s.AvailableQuantity += quantity
	s.ReservedQuantity -= quantity
	return nil
}

// Commit commits the reserved quantity (removes it completely)
func (s *Stock) Commit(quantity int32) error {
	if !s.CanCommit(quantity) {
		return errors.ErrInsufficientReservedStock
	}

	s.ReservedQuantity -= quantity
	return nil
}

// GetTotalQuantity returns the total quantity (available + reserved)
func (s *Stock) GetTotalQuantity() int32 {
	return s.AvailableQuantity + s.ReservedQuantity
}
