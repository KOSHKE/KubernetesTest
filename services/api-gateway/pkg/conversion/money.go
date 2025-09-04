package conversion

import "ecommerce-platform/pkg/common/valueobjects"

// PBMoney is a minimal interface satisfied by protobuf Money messages.
type PBMoney interface {
	GetAmount() int64
	GetCurrency() string
}

// MoneyFromPB converts a protobuf Money-like message to internal valueobjects.Money.
// Returns zero-value Money when m is nil.
func MoneyFromPB(m PBMoney) valueobjects.Money {
	if m == nil {
		return valueobjects.Money{}
	}
	return valueobjects.Money{Amount: m.GetAmount(), Currency: m.GetCurrency()}
}
