package order

import (
	realpb "github.com/kubernetestest/ecommerce-platform/proto-go/order"
)

// Re-export generated types into service-local import path to avoid module import issues.
type (
	Order                     = realpb.Order
	OrderItem                 = realpb.OrderItem
	Money                     = realpb.Money
	OrderStatus               = realpb.OrderStatus
	CreateOrderRequest        = realpb.CreateOrderRequest
	CreateOrderResponse       = realpb.CreateOrderResponse
	GetOrderRequest           = realpb.GetOrderRequest
	GetOrderResponse          = realpb.GetOrderResponse
	GetUserOrdersRequest      = realpb.GetUserOrdersRequest
	GetUserOrdersResponse     = realpb.GetUserOrdersResponse
	UpdateOrderStatusRequest  = realpb.UpdateOrderStatusRequest
	UpdateOrderStatusResponse = realpb.UpdateOrderStatusResponse
	CancelOrderRequest        = realpb.CancelOrderRequest
	CancelOrderResponse       = realpb.CancelOrderResponse
)

var (
	RegisterOrderServiceServer = realpb.RegisterOrderServiceServer

	OrderStatus_PENDING    = realpb.OrderStatus_PENDING
	OrderStatus_CONFIRMED  = realpb.OrderStatus_CONFIRMED
	OrderStatus_PROCESSING = realpb.OrderStatus_PROCESSING
	OrderStatus_SHIPPED    = realpb.OrderStatus_SHIPPED
	OrderStatus_DELIVERED  = realpb.OrderStatus_DELIVERED
	OrderStatus_CANCELLED  = realpb.OrderStatus_CANCELLED
)

type OrderServiceServer = realpb.OrderServiceServer

// Re-export the unimplemented server type for embedding
type UnimplementedOrderServiceServer = realpb.UnimplementedOrderServiceServer
