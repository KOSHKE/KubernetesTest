package conversion

import "github.com/kubernetestest/ecommerce-platform/services/api-gateway/pkg/types"

// PBMoney is a minimal interface satisfied by protobuf Money messages.
type PBMoney interface {
	GetAmount() int64
	GetCurrency() string
}

// MoneyFromPB converts a protobuf Money-like message to internal types.Money.
// Returns zero-value Money when m is nil.
func MoneyFromPB(m PBMoney) types.Money {
	if m == nil {
		return types.Money{}
	}
	return types.Money{Amount: m.GetAmount(), Currency: m.GetCurrency()}
}
