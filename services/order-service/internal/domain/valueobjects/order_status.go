package valueobjects

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusPending       OrderStatus = "PENDING"
	OrderStatusConfirmed     OrderStatus = "CONFIRMED"
	OrderStatusCancelled     OrderStatus = "CANCELLED"
	OrderStatusPaid          OrderStatus = "PAID"
	OrderStatusPaymentFailed OrderStatus = "PAYMENT_FAILED"
	OrderStatusStockReleased OrderStatus = "STOCK_RELEASED"
	OrderStatusCompleted     OrderStatus = "COMPLETED"
)

// String returns the string representation of the status
func (os OrderStatus) String() string {
	return string(os)
}
