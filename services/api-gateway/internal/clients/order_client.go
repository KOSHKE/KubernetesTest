package clients

import (
	"context"
	"time"

	"api-gateway/internal/types"
	orderpb "proto-go/order"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type OrderClient interface {
	Close() error
	CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error)
	GetOrder(ctx context.Context, orderID, userID string) (*Order, error)
	GetUserOrders(ctx context.Context, userID string, page, limit int32) ([]*Order, error)
	CancelOrder(ctx context.Context, orderID, userID string) (*Order, error)
}

type orderClient struct {
	conn   *grpc.ClientConn
	client orderpb.OrderServiceClient
}

type Order struct {
	ID              string      `json:"id"`
	UserID          string      `json:"user_id"`
	Status          string      `json:"status"`
	Items           []OrderItem `json:"items"`
	TotalAmount     types.Money `json:"total_amount"`
	ShippingAddress string      `json:"shipping_address"`
	CreatedAt       string      `json:"created_at"`
	UpdatedAt       string      `json:"updated_at"`
}

type OrderItem struct {
	ID          string      `json:"id"`
	ProductID   string      `json:"product_id"`
	ProductName string      `json:"product_name"`
	Quantity    int32       `json:"quantity"`
	Price       types.Money `json:"price"`
	Total       types.Money `json:"total"`
}

type CreateOrderRequest struct {
	UserID          string             `json:"user_id"`
	Items           []OrderItemRequest `json:"items"`
	ShippingAddress string             `json:"shipping_address"`
	Currency        string             `json:"currency"`
}

type OrderItemRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int32  `json:"quantity"`
}

func NewOrderClient(address string) (OrderClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, err
	}
	return &orderClient{conn: conn, client: orderpb.NewOrderServiceClient(conn)}, nil
}

func (c *orderClient) Close() error { return c.conn.Close() }

func (c *orderClient) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
	items := make([]*orderpb.OrderItemRequest, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, &orderpb.OrderItemRequest{ProductId: it.ProductID, Quantity: it.Quantity})
	}
	grpcReq := &orderpb.CreateOrderRequest{UserId: req.UserID, Items: items, ShippingAddress: req.ShippingAddress}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res, err := c.client.CreateOrder(ctx, grpcReq)
	if err != nil {
		return nil, err
	}
	return mapOrderFromPB(res.GetOrder()), nil
}

func (c *orderClient) GetOrder(ctx context.Context, orderID, userID string) (*Order, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res, err := c.client.GetOrder(ctx, &orderpb.GetOrderRequest{Id: orderID, UserId: userID})
	if err != nil {
		return nil, err
	}
	return mapOrderFromPB(res.GetOrder()), nil
}

func (c *orderClient) GetUserOrders(ctx context.Context, userID string, page, limit int32) ([]*Order, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res, err := c.client.GetUserOrders(ctx, &orderpb.GetUserOrdersRequest{UserId: userID, Page: page, Limit: limit})
	if err != nil {
		return nil, err
	}
	out := make([]*Order, 0, len(res.GetOrders()))
	for _, o := range res.GetOrders() {
		out = append(out, mapOrderFromPB(o))
	}
	return out, nil
}

func (c *orderClient) CancelOrder(ctx context.Context, orderID, userID string) (*Order, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res, err := c.client.CancelOrder(ctx, &orderpb.CancelOrderRequest{Id: orderID, UserId: userID})
	if err != nil {
		return nil, err
	}
	return mapOrderFromPB(res.GetOrder()), nil
}

// Mapping helpers from PB -> HTTP types
func mapOrderFromPB(o *orderpb.Order) *Order {
	if o == nil {
		return nil
	}
	items := make([]OrderItem, 0, len(o.Items))
	for _, it := range o.Items {
		items = append(items, OrderItem{
			ID:          it.Id,
			ProductID:   it.ProductId,
			ProductName: it.ProductName,
			Quantity:    it.Quantity,
			Price:       types.Money{Amount: it.Price.GetAmount(), Currency: it.Price.GetCurrency()},
			Total:       types.Money{Amount: it.Total.GetAmount(), Currency: it.Total.GetCurrency()},
		})
	}
	return &Order{
		ID:              o.Id,
		UserID:          o.UserId,
		Status:          mapStatusFromPB(o.Status),
		Items:           items,
		TotalAmount:     types.Money{Amount: o.TotalAmount.GetAmount(), Currency: o.TotalAmount.GetCurrency()},
		ShippingAddress: o.ShippingAddress,
		CreatedAt:       o.CreatedAt,
		UpdatedAt:       o.UpdatedAt,
	}
}

func mapStatusFromPB(s orderpb.OrderStatus) string {
	switch s {
	case orderpb.OrderStatus_PENDING:
		return "PENDING"
	case orderpb.OrderStatus_CONFIRMED:
		return "CONFIRMED"
	case orderpb.OrderStatus_PROCESSING:
		return "PROCESSING"
	case orderpb.OrderStatus_SHIPPED:
		return "SHIPPED"
	case orderpb.OrderStatus_DELIVERED:
		return "DELIVERED"
	case orderpb.OrderStatus_CANCELLED:
		return "CANCELLED"
	default:
		return "PENDING"
	}
}
