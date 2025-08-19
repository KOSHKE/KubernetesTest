package grpc

import (
	"context"
	"errors"

	appsvc "order-service/internal/app/services"
	derrors "order-service/internal/domain/errors"
	"order-service/internal/domain/models"
	orderpb "order-service/internal/pb/order"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PBOrderServer struct {
	orderpb.UnimplementedOrderServiceServer
	svc             *appsvc.OrderService
	defaultCurrency string
}

func NewPBOrderServer(svc *appsvc.OrderService, defaultCurrency string) *PBOrderServer {
	return &PBOrderServer{svc: svc, defaultCurrency: defaultCurrency}
}

func (s *PBOrderServer) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	// Validate basics
	if req.UserId == "" || req.ShippingAddress == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id and shipping_address are required")
	}
	if len(req.Items) == 0 {
		return nil, status.Error(codes.InvalidArgument, "items are required")
	}
	// Per-item validation
	for _, it := range req.Items {
		if it.ProductId == "" {
			return nil, status.Error(codes.InvalidArgument, "item.product_id is required")
		}
		if it.Quantity <= 0 {
			return nil, status.Error(codes.InvalidArgument, "item.quantity must be positive")
		}
	}
	// Map proto -> app request; product name/price not present in proto
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
		Currency:        s.defaultCurrency,
	})
	if err != nil {
		return nil, toStatusErr(err)
	}
	return &orderpb.CreateOrderResponse{Order: mapOrderToPB(order), Message: "Order created"}, nil
}

func (s *PBOrderServer) GetOrder(ctx context.Context, req *orderpb.GetOrderRequest) (*orderpb.GetOrderResponse, error) {
	ord, err := s.svc.GetOrder(ctx, req.Id, req.UserId)
	if err != nil {
		return nil, toStatusErr(err)
	}
	return &orderpb.GetOrderResponse{Order: mapOrderToPB(ord)}, nil
}

func (s *PBOrderServer) GetUserOrders(ctx context.Context, req *orderpb.GetUserOrdersRequest) (*orderpb.GetUserOrdersResponse, error) {
	orders, total, err := s.svc.GetUserOrders(ctx, req.UserId, int(req.Page), int(req.Limit))
	if err != nil {
		return nil, toStatusErr(err)
	}
	out := make([]*orderpb.Order, 0, len(orders))
	for _, o := range orders {
		out = append(out, mapOrderToPB(o))
	}
	return &orderpb.GetUserOrdersResponse{Orders: out, Total: int32(total)}, nil
}

func (s *PBOrderServer) UpdateOrderStatus(ctx context.Context, req *orderpb.UpdateOrderStatusRequest) (*orderpb.UpdateOrderStatusResponse, error) {
	var st models.OrderStatus
	switch req.Status {
	case orderpb.OrderStatus_PENDING:
		st = models.OrderStatusPending
	case orderpb.OrderStatus_CONFIRMED:
		st = models.OrderStatusConfirmed
	case orderpb.OrderStatus_PROCESSING:
		st = models.OrderStatusProcessing
	case orderpb.OrderStatus_SHIPPED:
		st = models.OrderStatusShipped
	case orderpb.OrderStatus_DELIVERED:
		st = models.OrderStatusDelivered
	case orderpb.OrderStatus_CANCELLED:
		st = models.OrderStatusCancelled
	default:
		return nil, status.Error(codes.InvalidArgument, "unknown order status")
	}
	ord, err := s.svc.UpdateOrderStatus(ctx, &appsvc.UpdateOrderStatusRequest{OrderID: req.Id, Status: st})
	if err != nil {
		return nil, toStatusErr(err)
	}
	return &orderpb.UpdateOrderStatusResponse{Order: mapOrderToPB(ord), Message: "Order status updated"}, nil
}

func (s *PBOrderServer) CancelOrder(ctx context.Context, req *orderpb.CancelOrderRequest) (*orderpb.CancelOrderResponse, error) {
	ord, err := s.svc.CancelOrder(ctx, req.Id, req.UserId)
	if err != nil {
		return nil, toStatusErr(err)
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
		CreatedAt:       timestamppb.New(o.CreatedAt),
		UpdatedAt:       timestamppb.New(o.UpdatedAt),
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

// toStatusErr maps domain/service errors to gRPC statuses
func toStatusErr(err error) error {
	switch {
	case errors.Is(err, derrors.ErrOrderNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, derrors.ErrOrderAccessDenied):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, derrors.ErrInvalidArgument):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
