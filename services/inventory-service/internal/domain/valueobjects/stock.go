package valueobjects

// Stock represents the stock information for a product as a value object
type Stock struct {
	AvailableQuantity int32
	ReservedQuantity  int32
}

// NewStock creates a new stock value object
func NewStock(availableQuantity, reservedQuantity int32) *Stock {
	return &Stock{
		AvailableQuantity: availableQuantity,
		ReservedQuantity:  reservedQuantity,
	}
}

// TotalQuantity returns the total quantity (available + reserved)
func (s *Stock) TotalQuantity() int32 {
	return s.AvailableQuantity + s.ReservedQuantity
}

// CanReserve checks if the requested quantity can be reserved
func (s *Stock) CanReserve(quantity int32) bool {
	return s.AvailableQuantity >= quantity
}

// Reserve reserves the specified quantity
func (s *Stock) Reserve(quantity int32) *Stock {
	if !s.CanReserve(quantity) {
		return s // Return unchanged if cannot reserve
	}

	return &Stock{
		AvailableQuantity: s.AvailableQuantity - quantity,
		ReservedQuantity:  s.ReservedQuantity + quantity,
	}
}

// Release releases the specified quantity from reserved back to available
func (s *Stock) Release(quantity int32) *Stock {
	if s.ReservedQuantity < quantity {
		quantity = s.ReservedQuantity // Release only what's available
	}

	return &Stock{
		AvailableQuantity: s.AvailableQuantity + quantity,
		ReservedQuantity:  s.ReservedQuantity - quantity,
	}
}

// Commit commits the reserved quantity (removes it completely)
func (s *Stock) Commit(quantity int32) *Stock {
	if s.ReservedQuantity < quantity {
		quantity = s.ReservedQuantity // Commit only what's available
	}

	return &Stock{
		AvailableQuantity: s.AvailableQuantity,
		ReservedQuantity:  s.ReservedQuantity - quantity,
	}
}

// AddAvailable adds quantity to available stock
func (s *Stock) AddAvailable(quantity int32) *Stock {
	return &Stock{
		AvailableQuantity: s.AvailableQuantity + quantity,
		ReservedQuantity:  s.ReservedQuantity,
	}
}

// Equals checks if two stock objects are equal
func (s *Stock) Equals(other *Stock) bool {
	if other == nil {
		return false
	}
	return s.AvailableQuantity == other.AvailableQuantity &&
		s.ReservedQuantity == other.ReservedQuantity
}
