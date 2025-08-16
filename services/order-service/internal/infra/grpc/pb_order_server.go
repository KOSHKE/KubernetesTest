package grpc

import (
	"context"

	appsvc "order-service/internal/app/services"
	"order-service/internal/domain/models"
	orderpb "order-service/internal/pb/order"
)

type PBOrderServer struct {
	orderpb.UnimplementedOrderServiceServer
	svc *appsvc.OrderService
}

func NewPBOrderServer(svc *appsvc.OrderService) *PBOrderServer { return &PBOrderServer{svc: svc} }

func (s *PBOrderServer) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	// Map proto -> app request (currency defaulted for now)
	items := make([]appsvc.OrderItemRequest, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, appsvc.OrderItemRequest{
			ProductID:   it.ProductId,
			ProductName: "",
			Quantity:    it.Quantity,
			Price:       0,
		})
	}
	order, err := s.svc.CreateOrder(ctx, &appsvc.CreateOrderRequest{
		UserID:          req.UserId,
		Items:           items,
		ShippingAddress: req.ShippingAddress,
		Currency:        "USD",
	})
	if err != nil {
		return nil, err
	}
	return &orderpb.CreateOrderResponse{Order: mapOrderToPB(order), Message: "Order created"}, nil
}

func (s *PBOrderServer) GetOrder(ctx context.Context, req *orderpb.GetOrderRequest) (*orderpb.GetOrderResponse, error) {
	ord, err := s.svc.GetOrder(ctx, req.Id, req.UserId)
	if err != nil {
		return nil, err
	}
	return &orderpb.GetOrderResponse{Order: mapOrderToPB(ord)}, nil
}

func (s *PBOrderServer) GetUserOrders(ctx context.Context, req *orderpb.GetUserOrdersRequest) (*orderpb.GetUserOrdersResponse, error) {
	orders, total, err := s.svc.GetUserOrders(ctx, req.UserId, int(req.Page), int(req.Limit))
	if err != nil {
		return nil, err
	}
	out := make([]*orderpb.Order, 0, len(orders))
	for _, o := range orders {
		out = append(out, mapOrderToPB(o))
	}
	return &orderpb.GetUserOrdersResponse{Orders: out, Total: int32(total)}, nil
}

func (s *PBOrderServer) UpdateOrderStatus(ctx context.Context, req *orderpb.UpdateOrderStatusRequest) (*orderpb.UpdateOrderStatusResponse, error) {
	ord, err := s.svc.UpdateOrderStatus(ctx, &appsvc.UpdateOrderStatusRequest{OrderID: req.Id, Status: mapStatusFromPB(req.Status)})
	if err != nil {
		return nil, err
	}
	return &orderpb.UpdateOrderStatusResponse{Order: mapOrderToPB(ord), Message: "Order status updated"}, nil
}

func (s *PBOrderServer) CancelOrder(ctx context.Context, req *orderpb.CancelOrderRequest) (*orderpb.CancelOrderResponse, error) {
	ord, err := s.svc.CancelOrder(ctx, req.Id, req.UserId)
	if err != nil {
		return nil, err
	}
	return &orderpb.CancelOrderResponse{Order: mapOrderToPB(ord), Message: "Order cancelled"}, nil
}

// Mapping helpers
func mapOrderToPB(o *models.Order) *orderpb.Order {
	items := make([]*orderpb.OrderItem, 0, len(o.Items))
	for _, it := range o.Items {
		items = append(items, &orderpb.OrderItem{
			Id:          it.ID,
			ProductId:   it.ProductID,
			ProductName: it.ProductName,
			Quantity:    it.Quantity,
			Price:       &orderpb.Money{Amount: it.Price, Currency: it.Currency},
			Total:       &orderpb.Money{Amount: it.Total, Currency: it.Currency},
		})
	}
	return &orderpb.Order{
		Id:              o.ID,
		UserId:          o.UserID,
		Status:          mapStatusToPB(o.Status),
		Items:           items,
		TotalAmount:     &orderpb.Money{Amount: o.TotalAmount, Currency: o.Currency},
		ShippingAddress: o.ShippingAddress,
		CreatedAt:       o.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       o.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func mapStatusToPB(s models.OrderStatus) orderpb.OrderStatus {
	switch s {
	case models.OrderStatusPending:
		return orderpb.OrderStatus_PENDING
	case models.OrderStatusConfirmed:
		return orderpb.OrderStatus_CONFIRMED
	case models.OrderStatusProcessing:
		return orderpb.OrderStatus_PROCESSING
	case models.OrderStatusShipped:
		return orderpb.OrderStatus_SHIPPED
	case models.OrderStatusDelivered:
		return orderpb.OrderStatus_DELIVERED
	case models.OrderStatusCancelled:
		return orderpb.OrderStatus_CANCELLED
	default:
		return orderpb.OrderStatus_PENDING
	}
}

func mapStatusFromPB(s orderpb.OrderStatus) models.OrderStatus {
	switch s {
	case orderpb.OrderStatus_PENDING:
		return models.OrderStatusPending
	case orderpb.OrderStatus_CONFIRMED:
		return models.OrderStatusConfirmed
	case orderpb.OrderStatus_PROCESSING:
		return models.OrderStatusProcessing
	case orderpb.OrderStatus_SHIPPED:
		return models.OrderStatusShipped
	case orderpb.OrderStatus_DELIVERED:
		return models.OrderStatusDelivered
	case orderpb.OrderStatus_CANCELLED:
		return models.OrderStatusCancelled
	default:
		return models.OrderStatusPending
	}
}
