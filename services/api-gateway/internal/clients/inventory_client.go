package clients

import (
	"context"
	"time"

	"api-gateway/internal/types"
	invpb "proto-go/inventory"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type InventoryClient interface {
	Close() error
	GetProducts(ctx context.Context, categoryID string, page, limit int32, search string) ([]*Product, int32, error)
	GetProduct(ctx context.Context, productID string) (*Product, error)
	GetCategories(ctx context.Context, activeOnly bool) ([]*Category, error)
	CheckStock(ctx context.Context, req *StockCheckRequest) (*StockCheckResponse, error)
}

type inventoryClient struct {
	conn   *grpc.ClientConn
	client invpb.InventoryServiceClient
}

type Product struct {
	ID            string      `json:"id"`
	Name          string      `json:"name"`
	Description   string      `json:"description"`
	Price         types.Money `json:"price"`
	CategoryID    string      `json:"category_id"`
	CategoryName  string      `json:"category_name"`
	StockQuantity int32       `json:"stock_quantity"`
	ImageURL      string      `json:"image_url"`
	IsActive      bool        `json:"is_active"`
	CreatedAt     string      `json:"created_at"`
	UpdatedAt     string      `json:"updated_at"`
}

type Category struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

type StockCheckRequest struct {
	Items []StockCheckItem `json:"items"`
}

type StockCheckItem struct {
	ProductID string `json:"product_id"`
	Quantity  int32  `json:"quantity"`
}

type StockCheckResponse struct {
	Results      []StockCheckResult `json:"results"`
	AllAvailable bool               `json:"all_available"`
}

type StockCheckResult struct {
	ProductID         string `json:"product_id"`
	RequestedQuantity int32  `json:"requested_quantity"`
	AvailableQuantity int32  `json:"available_quantity"`
	IsAvailable       bool   `json:"is_available"`
}

func NewInventoryClient(address string) (InventoryClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, err
	}

	return &inventoryClient{conn: conn, client: invpb.NewInventoryServiceClient(conn)}, nil
}

func (c *inventoryClient) Close() error { return c.conn.Close() }

func (c *inventoryClient) GetProducts(ctx context.Context, categoryID string, page, limit int32, search string) ([]*Product, int32, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res, err := c.client.GetProducts(ctx, &invpb.GetProductsRequest{CategoryId: categoryID, Page: page, Limit: limit, Search: search})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*Product, 0, len(res.GetProducts()))
	for _, p := range res.GetProducts() {
		out = append(out, mapProductFromPB(p))
	}
	return out, res.GetTotal(), nil
}

func (c *inventoryClient) GetProduct(ctx context.Context, productID string) (*Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res, err := c.client.GetProduct(ctx, &invpb.GetProductRequest{Id: productID})
	if err != nil {
		return nil, err
	}
	return mapProductFromPB(res.GetProduct()), nil
}

func (c *inventoryClient) GetCategories(ctx context.Context, activeOnly bool) ([]*Category, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res, err := c.client.GetCategories(ctx, &invpb.GetCategoriesRequest{ActiveOnly: activeOnly})
	if err != nil {
		return nil, err
	}
	out := make([]*Category, 0, len(res.GetCategories()))
	for _, cat := range res.GetCategories() {
		out = append(out, &Category{ID: cat.Id, Name: cat.Name, Description: cat.Description, IsActive: cat.IsActive})
	}
	return out, nil
}

func (c *inventoryClient) CheckStock(ctx context.Context, req *StockCheckRequest) (*StockCheckResponse, error) {
	grpcItems := make([]*invpb.StockCheckItem, 0, len(req.Items))
	for _, it := range req.Items {
		grpcItems = append(grpcItems, &invpb.StockCheckItem{ProductId: it.ProductID, Quantity: it.Quantity})
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	res, err := c.client.CheckStock(ctx, &invpb.CheckStockRequest{Items: grpcItems})
	if err != nil {
		return nil, err
	}
	results := make([]StockCheckResult, 0, len(res.GetResults()))
	for _, r := range res.GetResults() {
		results = append(results, StockCheckResult{
			ProductID:         r.ProductId,
			RequestedQuantity: r.RequestedQuantity,
			AvailableQuantity: r.AvailableQuantity,
			IsAvailable:       r.IsAvailable,
		})
	}
	return &StockCheckResponse{Results: results, AllAvailable: res.GetAllAvailable()}, nil
}

func mapProductFromPB(p *invpb.Product) *Product {
	if p == nil {
		return nil
	}
	return &Product{
		ID:            p.Id,
		Name:          p.Name,
		Description:   p.Description,
		Price:         types.Money{Amount: p.Price.GetAmount(), Currency: p.Price.GetCurrency()},
		CategoryID:    p.CategoryId,
		CategoryName:  p.CategoryName,
		StockQuantity: p.StockQuantity,
		ImageURL:      p.ImageUrl,
		IsActive:      p.IsActive,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}
