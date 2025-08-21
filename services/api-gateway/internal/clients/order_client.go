package clients

import (
	"context"

	"api-gateway/pkg/grpc"
	"api-gateway/pkg/types"
	orderpb "proto-go/order"
)

// ---------------- Order Client Interface ----------------

type OrderClient interface {
	Close() error
	CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error)
	GetOrder(ctx context.Context, orderID, userID string) (*Order, error)
	GetUserOrders(ctx context.Context, userID string, page, limit int32) ([]*Order, error)
}

type orderClient struct {
	*grpc.BaseClient
	client orderpb.OrderServiceClient
}

// ---------------- Order Models ----------------

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

// ---------------- Constructor ----------------

func NewOrderClient(address string) (OrderClient, error) {
	baseClient, err := grpc.NewBaseClient(address)
	if err != nil {
		return nil, err
	}
	return &orderClient{
		BaseClient: baseClient,
		client:     orderpb.NewOrderServiceClient(baseClient.GetConn()),
	}, nil
}

// ---------------- Order Methods ----------------

func (c *orderClient) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
	items := make([]*orderpb.OrderItemRequest, len(req.Items))
	for i, it := range req.Items {
		items[i] = &orderpb.OrderItemRequest{ProductId: it.ProductID, Quantity: it.Quantity}
	}
	grpcReq := &orderpb.CreateOrderRequest{
		UserId:          req.UserID,
		Items:           items,
		ShippingAddress: req.ShippingAddress,
	}

	resp, err := grpc.WithTimeoutResult(ctx, func(ctx context.Context) (*orderpb.CreateOrderResponse, error) {
		return c.client.CreateOrder(ctx, grpcReq)
	})
	if err != nil {
		return nil, err
	}
	return mapOrderFromPB(resp.GetOrder()), nil
}

func (c *orderClient) GetOrder(ctx context.Context, orderID, userID string) (*Order, error) {
	resp, err := grpc.WithTimeoutResult(ctx, func(ctx context.Context) (*orderpb.GetOrderResponse, error) {
		return c.client.GetOrder(ctx, &orderpb.GetOrderRequest{Id: orderID, UserId: userID})
	})
	if err != nil {
		return nil, err
	}
	return mapOrderFromPB(resp.GetOrder()), nil
}

func (c *orderClient) GetUserOrders(ctx context.Context, userID string, page, limit int32) ([]*Order, error) {
	resp, err := grpc.WithTimeoutResult(ctx, func(ctx context.Context) (*orderpb.GetUserOrdersResponse, error) {
		return c.client.GetUserOrders(ctx, &orderpb.GetUserOrdersRequest{
			UserId: userID, Page: page, Limit: limit,
		})
	})
	if err != nil {
		return nil, err
	}

	out := make([]*Order, len(resp.GetOrders()))
	for i, o := range resp.GetOrders() {
		out[i] = mapOrderFromPB(o)
	}
	return out, nil
}

// ---------------- Mapping Helpers ----------------

func mapMoneyFromPB(m *orderpb.Money) types.Money {
	if m == nil {
		return types.Money{}
	}
	return types.Money{Amount: m.GetAmount(), Currency: m.GetCurrency()}
}

func mapOrderItemFromPB(it *orderpb.OrderItem) OrderItem {
	if it == nil {
		return OrderItem{}
	}
	return OrderItem{
		ID:          it.Id,
		ProductID:   it.ProductId,
		ProductName: it.ProductName,
		Quantity:    it.Quantity,
		Price:       mapMoneyFromPB(it.Price),
		Total:       mapMoneyFromPB(it.Total),
	}
}

func mapOrderFromPB(o *orderpb.Order) *Order {
	if o == nil {
		return nil
	}
	items := make([]OrderItem, len(o.Items))
	for i, it := range o.Items {
		items[i] = mapOrderItemFromPB(it)
	}
	return &Order{
		ID:              o.Id,
		UserID:          o.UserId,
		Status:          mapStatusFromPB(o.Status),
		Items:           items,
		TotalAmount:     mapMoneyFromPB(o.TotalAmount),
		ShippingAddress: o.ShippingAddress,
		CreatedAt:       grpc.FormatTimestamp(o.CreatedAt),
		UpdatedAt:       grpc.FormatTimestamp(o.UpdatedAt),
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
		return "UNKNOWN"
	}
}
