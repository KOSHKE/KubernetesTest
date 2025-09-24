package clients

import (
	"context"
	"fmt"
	"time"

	dto "ecommerce-platform/pkg/common/dto/order-service"
	"ecommerce-platform/proto-go/order"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// OrderClient defines the interface for order operations
type OrderClient interface {
	Close() error
	CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error)
	GetOrder(ctx context.Context, orderID string) (*GetOrderResponse, error)
	GetUserOrders(ctx context.Context, userID string) (*GetUserOrdersResponse, error)
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

// CreateOrderRequest represents order creation request
type CreateOrderRequest = dto.CreateOrderRequest

// CreateOrderResponse represents order creation response
type CreateOrderResponse = dto.CreateOrderResponse

// GetOrderResponse represents get order response
type GetOrderResponse = dto.OrderResponse

// GetUserOrdersResponse represents get user orders response
type GetUserOrdersResponse = dto.OrdersListResponse

func (c *orderClient) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Convert to protobuf
	var items []*order.OrderItemRequest
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

	return &CreateOrderResponse{
		OrderID:     resp.Order.Id,
		TotalAmount: resp.Order.TotalAmount.Amount,
		Currency:    resp.Order.TotalAmount.Currency,
		Status:      resp.Order.Status.String(),
		Message:     resp.Message,
	}, nil
}

func (c *orderClient) GetOrder(ctx context.Context, orderID string) (*GetOrderResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.client.GetOrder(ctx, &order.GetOrderRequest{Id: orderID})
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Convert from protobuf
	var items []*dto.OrderItemResponse
	for _, item := range resp.Order.Items {
		items = append(items, &dto.OrderItemResponse{
			ProductID:   item.ProductId,
			ProductName: "", // TODO: get from product service
			Quantity:    item.Quantity,
			UnitPrice:   item.Price.Amount,
			TotalPrice:  item.Total.Amount,
		})
	}

	return &GetOrderResponse{
		ID:              resp.Order.Id,
		UserID:          resp.Order.UserId,
		Status:          resp.Order.Status.String(),
		Items:           items,
		ShippingAddress: dto.ShippingAddressDTO{Address: resp.Order.ShippingAddress},
		Currency:        resp.Order.TotalAmount.Currency,
		TotalAmount:     resp.Order.TotalAmount.Amount,
		CreatedAt:       resp.Order.CreatedAt.AsTime(),
		UpdatedAt:       resp.Order.UpdatedAt.AsTime(),
	}, nil
}

func (c *orderClient) GetUserOrders(ctx context.Context, userID string) (*GetUserOrdersResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.client.GetUserOrders(ctx, &order.GetUserOrdersRequest{UserId: userID})
	if err != nil {
		return nil, fmt.Errorf("failed to get user orders: %w", err)
	}

	// Convert from protobuf
	var orders []*dto.OrderResponse
	for _, orderItem := range resp.Orders {
		var items []*dto.OrderItemResponse
		for _, item := range orderItem.Items {
			items = append(items, &dto.OrderItemResponse{
				ProductID:   item.ProductId,
				ProductName: "", // TODO: get from product service
				Quantity:    item.Quantity,
				UnitPrice:   item.Price.Amount,
				TotalPrice:  item.Total.Amount,
			})
		}

		orders = append(orders, &dto.OrderResponse{
			ID:              orderItem.Id,
			UserID:          orderItem.UserId,
			Status:          orderItem.Status.String(),
			Items:           items,
			ShippingAddress: dto.ShippingAddressDTO{Address: orderItem.ShippingAddress},
			Currency:        orderItem.TotalAmount.Currency,
			TotalAmount:     orderItem.TotalAmount.Amount,
			CreatedAt:       orderItem.CreatedAt.AsTime(),
			UpdatedAt:       orderItem.UpdatedAt.AsTime(),
		})
	}

	return &GetUserOrdersResponse{
		Orders: orders,
		Total:  int64(len(orders)), // TODO: get from response
		Page:   1,                  // TODO: get from request
		Limit:  100,                // TODO: get from request
	}, nil
}
