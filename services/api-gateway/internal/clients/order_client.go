package clients

import (
	"context"
	"fmt"

	dto "ecommerce-platform/pkg/common/dto/order-service"
	commonpb "ecommerce-platform/proto-go/common"
	"ecommerce-platform/proto-go/order"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// OrderClient defines the interface for order operations
type OrderClient interface {
	Close() error
	// Caller should pass ctx with timeout or deadline to avoid hanging calls
	CreateOrder(ctx context.Context, req *dto.CreateOrderRequest) (*dto.CreateOrderResponse, error)
	// Caller should pass ctx with timeout or deadline to avoid hanging calls
	GetOrder(ctx context.Context, orderID string) (*dto.OrderResponse, error)
	GetUserOrders(ctx context.Context, userID string) ([]*dto.OrderResponse, error)
}

type orderClient struct {
	conn   *grpc.ClientConn
	client order.OrderServiceClient
}

// NewOrderClient creates a new gRPC client for order service
func NewOrderClient(address string) (OrderClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to order service: %w", err)
	}

	return &orderClient{
		conn:   conn,
		client: order.NewOrderServiceClient(conn),
	}, nil
}

func (c *orderClient) Close() error {
	return c.conn.Close()
}

// convertOrderItem maps protobuf OrderItem to DTO
func convertOrderItem(item *commonpb.OrderItem) *dto.OrderItemResponse {
	if item == nil {
		return nil
	}
	return &dto.OrderItemResponse{
		ProductID: item.ProductId,
		Quantity:  item.Quantity,
		UnitPrice: item.Price.Amount,
	}
}

// convertOrder maps protobuf Order to DTO
func convertOrder(orderPb *order.Order) *dto.OrderResponse {
	if orderPb == nil {
		return nil
	}

	items := make([]*dto.OrderItemResponse, 0, len(orderPb.Items))
	for _, it := range orderPb.Items {
		items = append(items, convertOrderItem(it))
	}

	return &dto.OrderResponse{
		ID:              orderPb.Id,
		UserID:          orderPb.UserId,
		Status:          orderPb.Status.String(),
		Items:           items,
		ShippingAddress: dto.ShippingAddressDTO{Address: orderPb.ShippingAddress},
		Currency:        orderPb.TotalAmount.Currency,
		TotalAmount:     orderPb.TotalAmount.Amount,
		CreatedAt:       orderPb.CreatedAt.AsTime(),
		UpdatedAt:       orderPb.UpdatedAt.AsTime(),
	}
}

// CreateOrder performs RPC call. Pass ctx with timeout or deadline in handlers.
func (c *orderClient) CreateOrder(ctx context.Context, req *dto.CreateOrderRequest) (*dto.CreateOrderResponse, error) {
	// Convert to protobuf
	items := make([]*order.OrderItemRequest, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, &order.OrderItemRequest{
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	pbReq := &order.CreateOrderRequest{
		UserId:          req.UserID,
		Items:           items,
		ShippingAddress: req.ShippingAddress.Address,
		Currency:        req.Currency,
	}

	resp, err := c.client.CreateOrder(ctx, pbReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	return &dto.CreateOrderResponse{
		OrderID:     resp.Order.Id,
		TotalAmount: resp.Order.TotalAmount.Amount,
		Currency:    resp.Order.TotalAmount.Currency,
		Status:      resp.Order.Status.String(),
		Message:     resp.Message,
	}, nil
}

// GetOrder performs RPC call. Pass ctx with timeout or deadline in handlers.
func (c *orderClient) GetOrder(ctx context.Context, orderID string) (*dto.OrderResponse, error) {
	resp, err := c.client.GetOrder(ctx, &order.GetOrderRequest{Id: orderID})
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Convert from protobuf
	return convertOrder(resp.Order), nil
}

func (c *orderClient) GetUserOrders(ctx context.Context, userID string) ([]*dto.OrderResponse, error) {
	resp, err := c.client.GetUserOrders(ctx, &order.GetUserOrdersRequest{UserId: userID})
	if err != nil {
		return nil, fmt.Errorf("failed to get user orders: %w", err)
	}

	// Convert from protobuf
	orders := make([]*dto.OrderResponse, 0, len(resp.Orders))
	for _, orderItem := range resp.Orders {
		orders = append(orders, convertOrder(orderItem))
	}

	return orders, nil
}
