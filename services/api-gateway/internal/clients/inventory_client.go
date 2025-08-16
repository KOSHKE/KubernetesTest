package clients

import (
	"context"
	"fmt"
	"time"

	"api-gateway/internal/types"

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
	conn *grpc.ClientConn
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

	return &inventoryClient{conn: conn}, nil
}

func (c *inventoryClient) Close() error { return c.conn.Close() }

func (c *inventoryClient) GetProducts(ctx context.Context, categoryID string, page, limit int32, search string) ([]*Product, int32, error) {
	products := []*Product{
		{ID: "prod-1", Name: "Wireless Headphones", Description: "High-quality wireless headphones with noise cancellation", Price: types.Money{Amount: 9999, Currency: "USD"}, CategoryID: "cat-1", CategoryName: "Electronics", StockQuantity: 50, ImageURL: "/images/headphones.jpg", IsActive: true, CreatedAt: time.Now().Add(-72 * time.Hour).Format(time.RFC3339), UpdatedAt: time.Now().Add(-24 * time.Hour).Format(time.RFC3339)},
		{ID: "prod-2", Name: "Smart Watch", Description: "Fitness tracking smart watch with heart rate monitor", Price: types.Money{Amount: 19999, Currency: "USD"}, CategoryID: "cat-1", CategoryName: "Electronics", StockQuantity: 25, ImageURL: "/images/smartwatch.jpg", IsActive: true, CreatedAt: time.Now().Add(-48 * time.Hour).Format(time.RFC3339), UpdatedAt: time.Now().Add(-12 * time.Hour).Format(time.RFC3339)},
		{ID: "prod-3", Name: "Coffee Mug", Description: "Ceramic coffee mug with custom design", Price: types.Money{Amount: 1599, Currency: "USD"}, CategoryID: "cat-2", CategoryName: "Home & Kitchen", StockQuantity: 100, ImageURL: "/images/mug.jpg", IsActive: true, CreatedAt: time.Now().Add(-24 * time.Hour).Format(time.RFC3339), UpdatedAt: time.Now().Add(-6 * time.Hour).Format(time.RFC3339)},
	}
	if categoryID != "" {
		var filtered []*Product
		for _, p := range products {
			if p.CategoryID == categoryID {
				filtered = append(filtered, p)
			}
		}
		products = filtered
	}
	if search != "" {
		var filtered []*Product
		for _, p := range products {
			if contains(p.Name, search) || contains(p.Description, search) {
				filtered = append(filtered, p)
			}
		}
		products = filtered
	}
	return products, int32(len(products)), nil
}

func (c *inventoryClient) GetProduct(ctx context.Context, productID string) (*Product, error) {
	return &Product{ID: productID, Name: "Mock Product " + productID, Description: "This is a mock product for testing purposes", Price: types.Money{Amount: 2999, Currency: "USD"}, CategoryID: "cat-1", CategoryName: "Electronics", StockQuantity: 10, ImageURL: fmt.Sprintf("/images/product-%s.jpg", productID), IsActive: true, CreatedAt: time.Now().Add(-24 * time.Hour).Format(time.RFC3339), UpdatedAt: time.Now().Format(time.RFC3339)}, nil
}

func (c *inventoryClient) GetCategories(ctx context.Context, activeOnly bool) ([]*Category, error) {
	categories := []*Category{{ID: "cat-1", Name: "Electronics", Description: "Electronic devices and gadgets", IsActive: true}, {ID: "cat-2", Name: "Home & Kitchen", Description: "Home and kitchen essentials", IsActive: true}, {ID: "cat-3", Name: "Books", Description: "Books and educational materials", IsActive: false}}
	if activeOnly {
		var filtered []*Category
		for _, c := range categories {
			if c.IsActive {
				filtered = append(filtered, c)
			}
		}
		return filtered, nil
	}
	return categories, nil
}

func (c *inventoryClient) CheckStock(ctx context.Context, req *StockCheckRequest) (*StockCheckResponse, error) {
	var results []StockCheckResult
	allAvailable := true
	for _, item := range req.Items {
		availableQuantity := int32(10)
		isAvailable := item.Quantity <= availableQuantity
		if !isAvailable {
			allAvailable = false
		}
		results = append(results, StockCheckResult{ProductID: item.ProductID, RequestedQuantity: item.Quantity, AvailableQuantity: availableQuantity, IsAvailable: isAvailable})
	}
	return &StockCheckResponse{Results: results, AllAvailable: allAvailable}, nil
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0)
}
