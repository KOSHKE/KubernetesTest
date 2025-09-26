package dtoconverters

import (
	orderService "ecommerce-platform/pkg/common/dto/order-service"
	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/services/order-service/internal/domain/aggregates"
	"ecommerce-platform/services/order-service/internal/domain/entities"
	orderValueObjects "ecommerce-platform/services/order-service/internal/domain/valueobjects"
)

// ToDomainShippingAddress converts ShippingAddressDTO to domain ShippingAddress
func ToDomainShippingAddress(s orderService.ShippingAddressDTO) (orderValueObjects.ShippingAddress, error) {
	return orderValueObjects.NewShippingAddress(s.Address)
}

// ToDomainCurrency converts string currency to domain Currency
func ToDomainCurrency(currency string) (valueobjects.Currency, error) {
	return valueobjects.NewCurrency(currency)
}

// ToDomainMoney converts int64 price to domain Money
func ToDomainMoney(price int64, currency string) (valueobjects.Money, error) {
	currencyObj, err := valueobjects.NewCurrency(currency)
	if err != nil {
		return valueobjects.Money{}, err
	}
	return valueobjects.Money{
		Amount:   price,
		Currency: currencyObj,
	}, nil
}

// ToDomainOrderStatus converts string status to domain OrderStatus
func ToDomainOrderStatus(status string) orderValueObjects.OrderStatus {
	return orderValueObjects.OrderStatus(status)
}

// NewOrderResponse creates an OrderResponse from Order aggregate
func NewOrderResponse(order *aggregates.Order) *orderService.OrderResponse {
	// Create items slice using the single item constructor
	items := make([]*orderService.OrderItemResponse, len(order.Items))
	for i, item := range order.Items {
		items[i] = NewOrderItemResponse(item)
	}

	return &orderService.OrderResponse{
		ID:              order.ID,
		UserID:          order.UserID,
		Status:          string(order.Status),
		Items:           items,
		ShippingAddress: ToShippingAddressDTO(order.ShippingAddress),
		Currency:        order.Currency.Code,
		TotalAmount:     order.TotalAmount.Amount,
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
	}
}

// NewOrderItemResponse creates an OrderItemResponse from domain entity
func NewOrderItemResponse(item *entities.OrderItem) *orderService.OrderItemResponse {
	return &orderService.OrderItemResponse{
		ProductID: item.ProductID,
		Quantity:  item.Quantity,
		UnitPrice: item.UnitPrice.Amount,
	}
}

// ToShippingAddressDTO converts domain ShippingAddress to ShippingAddressDTO
func ToShippingAddressDTO(addr orderValueObjects.ShippingAddress) orderService.ShippingAddressDTO {
	return orderService.ShippingAddressDTO{
		Address: addr.Value,
	}
}
