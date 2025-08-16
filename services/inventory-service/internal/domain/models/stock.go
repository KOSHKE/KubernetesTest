package models

import "fmt"

type Stock struct {
	ProductID         string
	AvailableQuantity int32
	ReservedQuantity  int32
}

func (s *Stock) CanReserve(qty int32) bool { return qty > 0 && s.AvailableQuantity >= qty }

func (s *Stock) Reserve(qty int32) error {
	if !s.CanReserve(qty) {
		return fmt.Errorf("insufficient stock")
	}
	s.AvailableQuantity -= qty
	s.ReservedQuantity += qty
	return nil
}

func (s *Stock) Release(qty int32) error {
	if qty <= 0 || s.ReservedQuantity < qty {
		return fmt.Errorf("invalid release qty")
	}
	s.ReservedQuantity -= qty
	s.AvailableQuantity += qty
	return nil
}
