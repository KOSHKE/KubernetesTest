package productinfoimpl

import (
	"context"
	"fmt"
	"time"

	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/ports/productinfo"
	invpb "github.com/kubernetestest/ecommerce-platform/proto-go/inventory"
)

// InventoryProvider implements productinfo.Provider using inventory-service gRPC client.
type InventoryProvider struct {
	client  invpb.InventoryServiceClient
	timeout time.Duration
}

func NewInventoryProvider(client invpb.InventoryServiceClient, timeout time.Duration) productinfo.Provider {
	return &InventoryProvider{client: client, timeout: timeout}
}

func (p *InventoryProvider) GetProduct(ctx context.Context, productID string) (*productinfo.ProductInfo, error) {
	// Determine timeout: use configured value or fallback to default
	timeout := p.timeout
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resp, err := p.client.GetProduct(ctx, &invpb.GetProductRequest{Id: productID})
	if err != nil {
		return nil, fmt.Errorf("inventory get product %s: %w", productID, err)
	}
	if resp.GetProduct() == nil {
		return nil, fmt.Errorf("inventory: product %s not found", productID)
	}
	pr := resp.GetProduct()
	return &productinfo.ProductInfo{
		Name:     pr.GetName(),
		Price:    pr.GetPrice().GetAmount(),
		Currency: pr.GetPrice().GetCurrency(),
	}, nil
}
