package grpc

import (
	"context"
	"errors"
	"time"

	orderpb "github.com/kubernetestest/ecommerce-platform/proto-go/order"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/application/dto"
	appsvc "github.com/kubernetestest/ecommerce-platform/services/order-service/internal/application/services"
	derrors "github.com/kubernetestest/ecommerce-platform/services/order-service/internal/domain/errors"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/domain/valueobjects"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/metrics"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PBOrderServer struct {
	orderpb.UnimplementedOrderServiceServer
	svc             *appsvc.OrderService
	defaultCurrency string
	metrics         metrics.OrderMetrics
}

func NewPBOrderServer(svc *appsvc.OrderService, defaultCurrency string, m metrics.OrderMetrics) *PBOrderServer {
	return &PBOrderServer{svc: svc, defaultCurrency: defaultCurrency, metrics: m}
}

func (s *PBOrderServer) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	start := time.Now()

	// Map proto -> app request
	items := make([]dto.OrderItemRequest, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, dto.OrderItemRequest{
			ProductID:   it.ProductId,
			ProductName: "",
			Quantity:    it.Quantity,
		})
	}

	order, err := s.svc.CreateOrder(ctx, &dto.CreateOrderRequest{
		UserID:          req.UserId,
		Items:           items,
		ShippingAddress: req.ShippingAddress,
		Currency:        s.defaultCurrency,
	})
	if err != nil {
		if s.metrics != nil {
			s.metrics.EventProcessed("order_creation", false)
			s.metrics.EventProcessingDuration(time.Since(start), "order_creation")
		}
		return nil, toStatusErr(err)
	}

	// Record metrics for successful request
	if s.metrics != nil {
		s.metrics.EventProcessed("order_creation", true)
		s.metrics.EventProcessingDuration(time.Since(start), "order_creation")
	}

	return &orderpb.CreateOrderResponse{Order: mapOrderResponseToPB(order), Message: "Order created"}, nil
}

func (s *PBOrderServer) GetOrder(ctx context.Context, req *orderpb.GetOrderRequest) (*orderpb.GetOrderResponse, error) {
	ord, err := s.svc.GetOrder(ctx, &dto.GetOrderRequest{OrderID: req.Id, UserID: req.UserId})
	if err != nil {
		return nil, toStatusErr(err)
	}
	return &orderpb.GetOrderResponse{Order: mapOrderResponseToPB(ord)}, nil
}

func (s *PBOrderServer) GetUserOrders(ctx context.Context, req *orderpb.GetUserOrdersRequest) (*orderpb.GetUserOrdersResponse, error) {
	response, err := s.svc.GetUserOrders(ctx, &dto.GetUserOrdersRequest{UserID: req.UserId, Page: int(req.Page), Limit: int(req.Limit)})
	if err != nil {
		return nil, toStatusErr(err)
	}
	out := make([]*orderpb.Order, 0, len(response.Orders))
	for _, o := range response.Orders {
		out = append(out, mapOrderResponseToPB(o))
	}
	return &orderpb.GetUserOrdersResponse{Orders: out, Total: int32(response.Total)}, nil
}

func (s *PBOrderServer) UpdateOrderStatus(ctx context.Context, req *orderpb.UpdateOrderStatusRequest) (*orderpb.UpdateOrderStatusResponse, error) {
	start := time.Now()

	var st valueobjects.OrderStatus
	switch req.Status {
	case orderpb.OrderStatus_PENDING:
		st = valueobjects.OrderStatusPending
	case orderpb.OrderStatus_CONFIRMED:
		st = valueobjects.OrderStatusConfirmed
	case orderpb.OrderStatus_PROCESSING:
		st = valueobjects.OrderStatusProcessing
	case orderpb.OrderStatus_SHIPPED:
		st = valueobjects.OrderStatusShipped
	case orderpb.OrderStatus_DELIVERED:
		st = valueobjects.OrderStatusDelivered
	case orderpb.OrderStatus_CANCELLED:
		st = valueobjects.OrderStatusCancelled
	default:
		return nil, status.Error(codes.InvalidArgument, "unknown order status")
	}
	ord, err := s.svc.UpdateOrderStatus(ctx, &dto.UpdateOrderStatusRequest{OrderID: req.Id, Status: st})
	if err != nil {
		// Record metrics for failed request
		if s.metrics != nil {
			s.metrics.EventProcessed("order_status_update", false)
			s.metrics.EventProcessingDuration(time.Since(start), "order_status_update")
		}
		return nil, toStatusErr(err)
	}

	// Record metrics for successful request
	if s.metrics != nil {
		s.metrics.EventProcessed("order_status_update", true)
		s.metrics.EventProcessingDuration(time.Since(start), "order_status_update")
	}

	return &orderpb.UpdateOrderStatusResponse{Order: mapOrderResponseToPB(ord), Message: "Order status updated"}, nil
}

func (s *PBOrderServer) CancelOrder(ctx context.Context, req *orderpb.CancelOrderRequest) (*orderpb.CancelOrderResponse, error) {
	ord, err := s.svc.CancelOrder(ctx, &dto.CancelOrderRequest{OrderID: req.Id, UserID: req.UserId})
	if err != nil {
		return nil, toStatusErr(err)
	}
	return &orderpb.CancelOrderResponse{Order: mapOrderResponseToPB(ord), Message: "Order cancelled"}, nil
}

// Mapping helpers
func mapOrderResponseToPB(o *dto.OrderResponse) *orderpb.Order {
	items := make([]*orderpb.OrderItem, 0, len(o.Items))
	for _, it := range o.Items {
		items = append(items, &orderpb.OrderItem{
			Id:          it.ProductID, // Using ProductID as ID since OrderItem doesn't have separate ID
			ProductId:   it.ProductID,
			ProductName: it.ProductName,
			Quantity:    it.Quantity,
			Price:       &orderpb.Money{Amount: it.UnitPrice.Amount, Currency: it.UnitPrice.Currency},
			Total:       &orderpb.Money{Amount: it.TotalPrice.Amount, Currency: it.TotalPrice.Currency},
		})
	}
	return &orderpb.Order{
		Id:              o.ID,
		UserId:          o.UserID,
		Status:          mapOrderStatusToPB(valueobjects.OrderStatus(o.Status)),
		Items:           items,
		TotalAmount:     &orderpb.Money{Amount: o.TotalAmount.Amount, Currency: o.TotalAmount.Currency},
		ShippingAddress: o.ShippingAddress,
		CreatedAt:       timestamppb.New(o.CreatedAt),
		UpdatedAt:       timestamppb.New(o.UpdatedAt),
	}
}

// mapOrderStatusToPB converts valueobjects.OrderStatus to protobuf OrderStatus
func mapOrderStatusToPB(s valueobjects.OrderStatus) orderpb.OrderStatus {
	switch s {
	case valueobjects.OrderStatusPending:
		return orderpb.OrderStatus_PENDING
	case valueobjects.OrderStatusConfirmed:
		return orderpb.OrderStatus_CONFIRMED
	case valueobjects.OrderStatusProcessing:
		return orderpb.OrderStatus_PROCESSING
	case valueobjects.OrderStatusShipped:
		return orderpb.OrderStatus_SHIPPED
	case valueobjects.OrderStatusDelivered:
		return orderpb.OrderStatus_DELIVERED
	case valueobjects.OrderStatusCancelled:
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
