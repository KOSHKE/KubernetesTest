package valueobjects

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "PENDING"
	PaymentStatusCompleted PaymentStatus = "COMPLETED"
	PaymentStatusFailed    PaymentStatus = "FAILED"
)

// String returns the string representation of the status
func (ps PaymentStatus) String() string {
	return string(ps)
}

// IsPending checks if status is pending
func (ps PaymentStatus) IsPending() bool {
	return ps == PaymentStatusPending
}

// IsCompleted checks if status is completed
func (ps PaymentStatus) IsCompleted() bool {
	return ps == PaymentStatusCompleted
}

// IsFailed checks if status is failed
func (ps PaymentStatus) IsFailed() bool {
	return ps == PaymentStatusFailed
}
