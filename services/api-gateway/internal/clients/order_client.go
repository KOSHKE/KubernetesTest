package clients

import (
	"context"
	"time"

	"api-gateway/internal/types"

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
	conn *grpc.ClientConn
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
	return &orderClient{conn: conn}, nil
}

func (c *orderClient) Close() error { return c.conn.Close() }

func (c *orderClient) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
	var items []OrderItem
	total := int64(0)
	for i, item := range req.Items {
		price := int64(2999)
		line := price * int64(item.Quantity)
		total += line
		items = append(items, OrderItem{ID: "item-" + string(rune(i+1)), ProductID: item.ProductID, ProductName: "Mock Product " + item.ProductID, Quantity: item.Quantity, Price: types.Money{Amount: price, Currency: req.Currency}, Total: types.Money{Amount: line, Currency: req.Currency}})
	}
	return &Order{ID: "order-" + time.Now().Format("20060102150405"), UserID: req.UserID, Status: "PENDING", Items: items, TotalAmount: types.Money{Amount: total, Currency: req.Currency}, ShippingAddress: req.ShippingAddress, CreatedAt: time.Now().Format(time.RFC3339), UpdatedAt: time.Now().Format(time.RFC3339)}, nil
}

func (c *orderClient) GetOrder(ctx context.Context, orderID, userID string) (*Order, error) {
	return &Order{ID: orderID, UserID: userID, Status: "CONFIRMED", Items: []OrderItem{{ID: "item-1", ProductID: "prod-1", ProductName: "Mock Product 1", Quantity: 2, Price: types.Money{Amount: 2999, Currency: "USD"}, Total: types.Money{Amount: 5998, Currency: "USD"}}}, TotalAmount: types.Money{Amount: 5998, Currency: "USD"}, ShippingAddress: "123 Main St, City, Country", CreatedAt: time.Now().Add(-2 * time.Hour).Format(time.RFC3339), UpdatedAt: time.Now().Add(-1 * time.Hour).Format(time.RFC3339)}, nil
}

func (c *orderClient) GetUserOrders(ctx context.Context, userID string, page, limit int32) ([]*Order, error) {
	orders := []*Order{{ID: "order-1", UserID: userID, Status: "DELIVERED", Items: []OrderItem{{ID: "item-1", ProductID: "prod-1", ProductName: "Mock Product 1", Quantity: 1, Price: types.Money{Amount: 2999, Currency: "USD"}, Total: types.Money{Amount: 2999, Currency: "USD"}}}, TotalAmount: types.Money{Amount: 2999, Currency: "USD"}, ShippingAddress: "123 Main St, City, Country", CreatedAt: time.Now().Add(-48 * time.Hour).Format(time.RFC3339), UpdatedAt: time.Now().Add(-24 * time.Hour).Format(time.RFC3339)}, {ID: "order-2", UserID: userID, Status: "PROCESSING", Items: []OrderItem{{ID: "item-2", ProductID: "prod-2", ProductName: "Mock Product 2", Quantity: 3, Price: types.Money{Amount: 1999, Currency: "USD"}, Total: types.Money{Amount: 5997, Currency: "USD"}}}, TotalAmount: types.Money{Amount: 5997, Currency: "USD"}, ShippingAddress: "456 Oak Ave, City, Country", CreatedAt: time.Now().Add(-12 * time.Hour).Format(time.RFC3339), UpdatedAt: time.Now().Add(-6 * time.Hour).Format(time.RFC3339)}}
	return orders, nil
}

func (c *orderClient) CancelOrder(ctx context.Context, orderID, userID string) (*Order, error) {
	order, _ := c.GetOrder(ctx, orderID, userID)
	order.Status = "CANCELLED"
	order.UpdatedAt = time.Now().Format(time.RFC3339)
	return order, nil
}
