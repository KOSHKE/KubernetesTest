package grpc

import (
	"context"
	stdErrors "errors"
	"time"

	dto "ecommerce-platform/pkg/common/dto/order-service"
	"ecommerce-platform/pkg/common/grpcutils"
	"ecommerce-platform/proto-go/common"
	orderpb "ecommerce-platform/proto-go/order"
	appsvc "ecommerce-platform/services/order-service/internal/application/services"
	orderValueObjects "ecommerce-platform/services/order-service/internal/domain/valueobjects"
	"ecommerce-platform/services/order-service/internal/metrics"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type PBOrderServer struct {
	orderpb.UnimplementedOrderServiceServer
	appService *appsvc.OrderApplicationService
	metrics    metrics.OrderMetrics
}

func NewPBOrderServer(appService *appsvc.OrderApplicationService, metrics metrics.OrderMetrics) *PBOrderServer {
	return &PBOrderServer{
		appService: appService,
		metrics:    metrics,
	}
}

func (s *PBOrderServer) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	start := time.Now()
	method := "CreateOrder"
	status := "success"
	defer func() {
		s.metrics.GRPCRequestDuration(method, time.Since(start))
		s.metrics.GRPCRequestTotal(method, status)
	}()

	items := make([]dto.OrderItemRequest, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, dto.OrderItemRequest{
			ProductID:   it.ProductId,
			ProductName: it.ProductName,
			Quantity:    it.Quantity,
			Price:       it.UnitPrice,
		})
	}

	// Parse shipping address
	shippingAddr := dto.ShippingAddressDTO{
		Address: req.ShippingAddress,
	}

	order, err := s.appService.CreateOrder(ctx, &dto.CreateOrderRequest{
		UserID:          req.UserId,
		Items:           items,
		ShippingAddress: shippingAddr,
		Currency:        req.Currency,
	})
	if err != nil {
		status = "error"
		return nil, grpcutils.MapErrorToStatus(err)
	}

	return &orderpb.CreateOrderResponse{Order: mapOrderResponseToPB(order)}, nil
}

func (s *PBOrderServer) GetOrder(ctx context.Context, req *orderpb.GetOrderRequest) (*orderpb.GetOrderResponse, error) {
	start := time.Now()
	method := "GetOrder"
	status := "success"
	defer func() {
		s.metrics.GRPCRequestDuration(method, time.Since(start))
		s.metrics.GRPCRequestTotal(method, status)
	}()

	ord, err := s.appService.GetOrder(ctx, &dto.GetOrderRequest{OrderID: req.Id, UserID: req.UserId})
	if err != nil {
		status = "error"
		return nil, grpcutils.MapErrorToStatus(err)
	}
	return &orderpb.GetOrderResponse{Order: mapOrderResponseToPB(ord)}, nil
}

func (s *PBOrderServer) GetUserOrders(ctx context.Context, req *orderpb.GetUserOrdersRequest) (*orderpb.GetUserOrdersResponse, error) {
	start := time.Now()
	method := "GetUserOrders"
	status := "success"
	defer func() {
		s.metrics.GRPCRequestDuration(method, time.Since(start))
		s.metrics.GRPCRequestTotal(method, status)
	}()

	response, err := s.appService.GetUserOrders(ctx, &dto.GetUserOrdersRequest{UserID: req.UserId, Page: int(req.Page), Limit: int(req.Limit)})
	if err != nil {
		status = "error"
		return nil, grpcutils.MapErrorToStatus(err)
	}
	out := make([]*orderpb.Order, 0, len(response.Orders))
	for _, o := range response.Orders {
		out = append(out, mapOrderResponseToPB(o))
	}
	return &orderpb.GetUserOrdersResponse{Orders: out, Total: int32(response.Total)}, nil
}

func (s *PBOrderServer) UpdateOrderStatus(ctx context.Context, req *orderpb.UpdateOrderStatusRequest) (*orderpb.UpdateOrderStatusResponse, error) {
	start := time.Now()
	method := "UpdateOrderStatus"
	status := "success"
	defer func() {
		s.metrics.GRPCRequestDuration(method, time.Since(start))
		s.metrics.GRPCRequestTotal(method, status)
	}()

	var st orderValueObjects.OrderStatus
	switch req.Status {
	case orderpb.OrderStatus_PENDING:
		st = orderValueObjects.OrderStatusPending
	case orderpb.OrderStatus_CONFIRMED:
		st = orderValueObjects.OrderStatusConfirmed
	case orderpb.OrderStatus_CANCELLED:
		st = orderValueObjects.OrderStatusCancelled
	default:
		status = "error"
		return nil, grpcutils.MapErrorToStatus(stdErrors.New("unknown order status"))
	}
	ord, err := s.appService.UpdateOrderStatus(ctx, &dto.UpdateOrderStatusRequest{OrderID: req.Id, Status: string(st)})
	if err != nil {
		status = "error"
		return nil, grpcutils.MapErrorToStatus(err)
	}

	return &orderpb.UpdateOrderStatusResponse{Order: mapOrderResponseToPB(ord)}, nil
}

func (s *PBOrderServer) CancelOrder(ctx context.Context, req *orderpb.CancelOrderRequest) (*orderpb.CancelOrderResponse, error) {
	start := time.Now()
	method := "CancelOrder"
	status := "success"
	defer func() {
		s.metrics.GRPCRequestDuration(method, time.Since(start))
		s.metrics.GRPCRequestTotal(method, status)
	}()

	ord, err := s.appService.CancelOrder(ctx, &dto.CancelOrderRequest{OrderID: req.Id, UserID: req.UserId})
	if err != nil {
		status = "error"
		return nil, grpcutils.MapErrorToStatus(err)
	}
	return &orderpb.CancelOrderResponse{Order: mapOrderResponseToPB(ord)}, nil
}

// Mapping helpers
func mapOrderResponseToPB(o *dto.OrderResponse) *orderpb.Order {
	items := make([]*common.OrderItem, 0, len(o.Items))
	for _, it := range o.Items {
		items = append(items, &common.OrderItem{
			ProductId: it.ProductID,
			Quantity:  it.Quantity,
			Price:     &common.Money{Amount: it.UnitPrice, Currency: o.Currency},
		})
	}
	return &orderpb.Order{
		Id:              o.ID,
		UserId:          o.UserID,
		Status:          mapOrderStatusToPB(orderValueObjects.OrderStatus(o.Status)),
		Items:           items,
		TotalAmount:     &common.Money{Amount: o.TotalAmount, Currency: o.Currency},
		ShippingAddress: o.ShippingAddress.Address,
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
