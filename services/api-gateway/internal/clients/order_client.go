package clients

import (
	"context"
	"fmt"
	"time"

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
type CreateOrderRequest struct {
	UserID string `json:"user_id"`
	Items  []Item `json:"items"`
}

type Item struct {
	ProductID string `json:"product_id"`
	Quantity  int32  `json:"quantity"`
}

// CreateOrderResponse represents order creation response
type CreateOrderResponse struct {
	OrderID string `json:"order_id"`
	Message string `json:"message"`
}

// GetOrderResponse represents get order response
type GetOrderResponse struct {
	OrderID string `json:"order_id"`
	UserID  string `json:"user_id"`
	Status  string `json:"status"`
	Items   []Item `json:"items"`
}

// GetUserOrdersResponse represents get user orders response
type GetUserOrdersResponse struct {
	Orders []GetOrderResponse `json:"orders"`
}

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
		UserId: req.UserID,
		Items:  items,
	}

	resp, err := c.client.CreateOrder(ctx, pbReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	return &CreateOrderResponse{
		OrderID: resp.Order.Id,
		Message: resp.Message,
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
	var items []Item
	for _, item := range resp.Order.Items {
		items = append(items, Item{
			ProductID: item.ProductId,
			Quantity:  item.Quantity,
		})
	}

	return &GetOrderResponse{
		OrderID: resp.Order.Id,
		UserID:  resp.Order.UserId,
		Status:  resp.Order.Status.String(),
		Items:   items,
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
	var orders []GetOrderResponse
	for _, orderItem := range resp.Orders {
		var items []Item
		for _, item := range orderItem.Items {
			items = append(items, Item{
				ProductID: item.ProductId,
				Quantity:  item.Quantity,
			})
		}

		orders = append(orders, GetOrderResponse{
			OrderID: orderItem.Id,
			UserID:  orderItem.UserId,
			Status:  orderItem.Status.String(),
			Items:   items,
		})
	}

	return &GetUserOrdersResponse{
		Orders: orders,
	}, nil
}
