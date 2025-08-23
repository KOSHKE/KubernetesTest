package clients

import (
	"context"

	"github.com/kubernetestest/ecommerce-platform/services/api-gateway/pkg/conversion"
	"github.com/kubernetestest/ecommerce-platform/services/api-gateway/pkg/grpc"
	"github.com/kubernetestest/ecommerce-platform/services/api-gateway/pkg/types"
	invpb "github.com/kubernetestest/ecommerce-platform/proto-go/inventory"
)

// ---------------- Inventory Client Interface ----------------

type InventoryClient interface {
	Close() error
	GetProducts(ctx context.Context, categoryID string, page, limit int32, search string) ([]*Product, int32, error)
	GetProduct(ctx context.Context, productID string) (*Product, error)
	GetCategories(ctx context.Context, activeOnly bool) ([]*Category, error)
	CheckStock(ctx context.Context, req *StockCheckRequest) (*StockCheckResponse, error)
}

type inventoryClient struct {
	*grpc.BaseClient
	client invpb.InventoryServiceClient
}

// ---------------- Inventory Models ----------------

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

// ---------------- Constructor ----------------

func NewInventoryClient(address string) (InventoryClient, error) {
	baseClient, err := grpc.NewBaseClient(address)
	if err != nil {
		return nil, err
	}
	return &inventoryClient{
		BaseClient: baseClient,
		client:     invpb.NewInventoryServiceClient(baseClient.GetConn()),
	}, nil
}

// ---------------- Inventory Methods ----------------

func (c *inventoryClient) GetProducts(ctx context.Context, categoryID string, page, limit int32, search string) ([]*Product, int32, error) {
	resp, err := grpc.WithTimeoutResult(ctx, func(ctx context.Context) (*invpb.GetProductsResponse, error) {
		return c.client.GetProducts(ctx, &invpb.GetProductsRequest{
			CategoryId: categoryID,
			Page:       page,
			Limit:      limit,
			Search:     search,
		})
	})
	if err != nil {
		return nil, 0, err
	}

	out := make([]*Product, len(resp.GetProducts()))
	for i, p := range resp.GetProducts() {
		out[i] = mapProductFromPB(p)
	}
	return out, resp.GetTotal(), nil
}

func (c *inventoryClient) GetProduct(ctx context.Context, productID string) (*Product, error) {
	resp, err := grpc.WithTimeoutResult(ctx, func(ctx context.Context) (*invpb.GetProductResponse, error) {
		return c.client.GetProduct(ctx, &invpb.GetProductRequest{Id: productID})
	})
	if err != nil {
		return nil, err
	}
	return mapProductFromPB(resp.GetProduct()), nil
}

func (c *inventoryClient) GetCategories(ctx context.Context, activeOnly bool) ([]*Category, error) {
	resp, err := grpc.WithTimeoutResult(ctx, func(ctx context.Context) (*invpb.GetCategoriesResponse, error) {
		return c.client.GetCategories(ctx, &invpb.GetCategoriesRequest{ActiveOnly: activeOnly})
	})
	if err != nil {
		return nil, err
	}

	out := make([]*Category, len(resp.GetCategories()))
	for i, cat := range resp.GetCategories() {
		out[i] = mapCategoryFromPB(cat)
	}
	return out, nil
}

func (c *inventoryClient) CheckStock(ctx context.Context, req *StockCheckRequest) (*StockCheckResponse, error) {
	grpcItems := make([]*invpb.StockCheckItem, len(req.Items))
	for i, it := range req.Items {
		grpcItems[i] = &invpb.StockCheckItem{ProductId: it.ProductID, Quantity: it.Quantity}
	}

	resp, err := grpc.WithTimeoutResult(ctx, func(ctx context.Context) (*invpb.CheckStockResponse, error) {
		return c.client.CheckStock(ctx, &invpb.CheckStockRequest{Items: grpcItems})
	})
	if err != nil {
		return nil, err
	}

	results := make([]StockCheckResult, len(resp.GetResults()))
	for i, r := range resp.GetResults() {
		results[i] = mapStockCheckResultFromPB(r)
	}

	return &StockCheckResponse{Results: results, AllAvailable: resp.GetAllAvailable()}, nil
}

// ---------------- Mapping Helpers ----------------

func mapProductFromPB(p *invpb.Product) *Product {
	if p == nil {
		return nil
	}
	return &Product{
		ID:            p.Id,
		Name:          p.Name,
		Description:   p.Description,
		Price:         conversion.MoneyFromPB(p.Price),
		CategoryID:    p.CategoryId,
		CategoryName:  p.CategoryName,
		StockQuantity: p.StockQuantity,
		ImageURL:      p.ImageUrl,
		IsActive:      p.IsActive,
		CreatedAt:     grpc.FormatTimestamp(p.CreatedAt),
		UpdatedAt:     grpc.FormatTimestamp(p.UpdatedAt),
	}
}

func mapCategoryFromPB(c *invpb.Category) *Category {
	return &Category{
		ID:          c.Id,
		Name:        c.Name,
		Description: c.Description,
		IsActive:    c.IsActive,
	}
}

func mapStockCheckResultFromPB(r *invpb.StockCheckResult) StockCheckResult {
	return StockCheckResult{
		ProductID:         r.ProductId,
		RequestedQuantity: r.RequestedQuantity,
		AvailableQuantity: r.AvailableQuantity,
		IsAvailable:       r.IsAvailable,
	}
}
