package clients

import (
	"context"
	"fmt"

	"ecommerce-platform/proto-go/inventory"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// InventoryClient defines the interface for inventory operations used by the gateway
type InventoryClient interface {
	Close() error
	// Caller should pass ctx with timeout or deadline to avoid hanging calls
	GetProducts(ctx context.Context, categoryID string, page, limit int, search string) (*ListProductsResult, error)
	// Caller should pass ctx with timeout or deadline to avoid hanging calls
	GetProduct(ctx context.Context, id string) (*Product, error)
}

type inventoryClient struct {
	conn   *grpc.ClientConn
	client inventory.InventoryServiceClient
}

// NewInventoryClient creates a new gRPC client for inventory service
func NewInventoryClient(address string) (InventoryClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to inventory service: %w", err)
	}

	return &inventoryClient{
		conn:   conn,
		client: inventory.NewInventoryServiceClient(conn),
	}, nil
}

func (c *inventoryClient) Close() error { return c.conn.Close() }

// Types returned to handlers
type Product struct {
	ID            string
	Name          string
	PriceAmount   int64
	Currency      string
	ImageURL      string
	StockQuantity int32
}

type ListProductsResult struct {
	Products []*Product
	Total    int32
	Page     int
	Limit    int
}

func (c *inventoryClient) GetProducts(ctx context.Context, categoryID string, page, limit int, search string) (*ListProductsResult, error) {
	req := &inventory.GetProductsRequest{
		CategoryId: categoryID,
		Page:       int32(page),
		Limit:      int32(limit),
		Search:     search,
	}

	resp, err := c.client.GetProducts(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	products := make([]*Product, 0, len(resp.Products))
	for _, p := range resp.Products {
		products = append(products, &Product{
			ID:            p.Id,
			Name:          p.Name,
			PriceAmount:   p.Price.Amount,
			Currency:      p.Price.Currency,
			ImageURL:      p.ImageUrl,
			StockQuantity: p.StockQuantity,
		})
	}

	return &ListProductsResult{
		Products: products,
		Total:    resp.Total,
		Page:     page,
		Limit:    limit,
	}, nil
}

func (c *inventoryClient) GetProduct(ctx context.Context, id string) (*Product, error) {
	resp, err := c.client.GetProduct(ctx, &inventory.GetProductRequest{Id: id})
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	p := resp.GetProduct()
	return &Product{
		ID:            p.GetId(),
		Name:          p.GetName(),
		PriceAmount:   p.GetPrice().GetAmount(),
		Currency:      p.GetPrice().GetCurrency(),
		ImageURL:      p.GetImageUrl(),
		StockQuantity: p.GetStockQuantity(),
	}, nil
}
