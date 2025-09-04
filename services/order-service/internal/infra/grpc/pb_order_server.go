package grpc

import (
	"context"
	stdErrors "errors"

	"ecommerce-platform/pkg/common/grpcutils"
	orderpb "ecommerce-platform/proto-go/order"
	"ecommerce-platform/services/order-service/internal/application/dto"
	appsvc "ecommerce-platform/services/order-service/internal/application/services"
	orderValueObjects "ecommerce-platform/services/order-service/internal/domain/valueobjects"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type PBOrderServer struct {
	orderpb.UnimplementedOrderServiceServer
	svc *appsvc.OrderApplicationService
}

func NewPBOrderServer(svc *appsvc.OrderApplicationService) *PBOrderServer {
	return &PBOrderServer{svc: svc}
}

func (s *PBOrderServer) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	// Map proto -> app request
	items := make([]dto.OrderItemRequest, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, dto.OrderItemRequest{
			ProductID: it.ProductId,
			Quantity:  it.Quantity,
		})
	}

	order, err := s.svc.CreateOrder(ctx, &dto.CreateOrderRequest{
		UserID:          req.UserId,
		Items:           items,
		ShippingAddress: req.ShippingAddress,
		Currency:        req.Currency,
	})
	if err != nil {
		return nil, grpcutils.MapErrorToStatus(err)
	}

	return &orderpb.CreateOrderResponse{Order: mapOrderResponseToPB(order)}, nil
}

func (s *PBOrderServer) GetOrder(ctx context.Context, req *orderpb.GetOrderRequest) (*orderpb.GetOrderResponse, error) {
	ord, err := s.svc.GetOrder(ctx, &dto.GetOrderRequest{OrderID: req.Id, UserID: req.UserId})
	if err != nil {
		return nil, grpcutils.MapErrorToStatus(err)
	}
	return &orderpb.GetOrderResponse{Order: mapOrderResponseToPB(ord)}, nil
}

func (s *PBOrderServer) GetUserOrders(ctx context.Context, req *orderpb.GetUserOrdersRequest) (*orderpb.GetUserOrdersResponse, error) {
	response, err := s.svc.GetUserOrders(ctx, &dto.GetUserOrdersRequest{UserID: req.UserId, Page: int(req.Page), Limit: int(req.Limit)})
	if err != nil {
		return nil, grpcutils.MapErrorToStatus(err)
	}
	out := make([]*orderpb.Order, 0, len(response.Orders))
	for _, o := range response.Orders {
		out = append(out, mapOrderResponseToPB(o))
	}
	return &orderpb.GetUserOrdersResponse{Orders: out, Total: int32(response.Total)}, nil
}

func (s *PBOrderServer) UpdateOrderStatus(ctx context.Context, req *orderpb.UpdateOrderStatusRequest) (*orderpb.UpdateOrderStatusResponse, error) {
	var st orderValueObjects.OrderStatus
	switch req.Status {
	case orderpb.OrderStatus_PENDING:
		st = orderValueObjects.OrderStatusPending
	case orderpb.OrderStatus_CONFIRMED:
		st = orderValueObjects.OrderStatusConfirmed
	case orderpb.OrderStatus_CANCELLED:
		st = orderValueObjects.OrderStatusCancelled
	default:
		return nil, grpcutils.MapErrorToStatus(stdErrors.New("unknown order status"))
	}
	ord, err := s.svc.UpdateOrderStatus(ctx, &dto.UpdateOrderStatusRequest{OrderID: req.Id, Status: st})
	if err != nil {
		return nil, grpcutils.MapErrorToStatus(err)
	}

	return &orderpb.UpdateOrderStatusResponse{Order: mapOrderResponseToPB(ord)}, nil
}

func (s *PBOrderServer) CancelOrder(ctx context.Context, req *orderpb.CancelOrderRequest) (*orderpb.CancelOrderResponse, error) {
	ord, err := s.svc.CancelOrder(ctx, &dto.CancelOrderRequest{OrderID: req.Id, UserID: req.UserId})
	if err != nil {
		return nil, grpcutils.MapErrorToStatus(err)
	}
	return &orderpb.CancelOrderResponse{Order: mapOrderResponseToPB(ord)}, nil
}

// Mapping helpers
func mapOrderResponseToPB(o *dto.OrderResponse) *orderpb.Order {
	items := make([]*orderpb.OrderItem, 0, len(o.Items))
	for _, it := range o.Items {
		items = append(items, &orderpb.OrderItem{
			ProductId:   it.ProductID,
			ProductName: it.ProductName,
			Quantity:    it.Quantity,
			Price:       &orderpb.Money{Amount: it.UnitPrice.Amount, Currency: it.UnitPrice.Currency.Code()},
			Total:       &orderpb.Money{Amount: it.TotalPrice.Amount, Currency: it.TotalPrice.Currency.Code()},
		})
	}
	return &orderpb.Order{
		Id:              o.ID,
		UserId:          o.UserID,
		Status:          mapOrderStatusToPB(orderValueObjects.OrderStatus(o.Status)),
		Items:           items,
		TotalAmount:     &orderpb.Money{Amount: o.TotalAmount.Amount, Currency: o.TotalAmount.Currency.Code()},
		ShippingAddress: o.ShippingAddress,
		CreatedAt:       timestamppb.New(o.CreatedAt),
		UpdatedAt:       timestamppb.New(o.UpdatedAt),
	}
}

// mapOrderStatusToPB converts orderValueObjects.OrderStatus to protobuf OrderStatus
func mapOrderStatusToPB(s orderValueObjects.OrderStatus) orderpb.OrderStatus {
	switch s {
	case orderValueObjects.OrderStatusPending:
		return orderpb.OrderStatus_PENDING
	case orderValueObjects.OrderStatusConfirmed:
		return orderpb.OrderStatus_CONFIRMED
	case orderValueObjects.OrderStatusCancelled:
		return orderpb.OrderStatus_CANCELLED
	default:
		return orderpb.OrderStatus_PENDING
	}
}
