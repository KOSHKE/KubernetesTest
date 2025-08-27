package productinfoimpl

import (
	"context"
	"fmt"
	"time"

	invpb "github.com/kubernetestest/ecommerce-platform/proto-go/inventory"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/domain/errors"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/domain/ports/productinfo"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/domain/valueobjects"
)

// InventoryProvider implements productinfo.Provider using inventory-service gRPC client.
type InventoryProvider struct {
	client  invpb.InventoryServiceClient
	timeout time.Duration
}

// NewInventoryProvider creates a new InventoryProvider instance.
// timeout must be > 0, otherwise a default 3-second timeout will be used.
func NewInventoryProvider(client invpb.InventoryServiceClient, timeout time.Duration) productinfo.Provider {
	return &InventoryProvider{client: client, timeout: timeout}
}

func (p *InventoryProvider) GetProduct(ctx context.Context, productID string) (*productinfo.ProductInfo, error) {
	// Use fallback timeout if provided timeout is <= 0
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
		return nil, fmt.Errorf("%w: %s", errors.ErrProductNotFound, productID)
	}
	pr := resp.GetProduct()

	// Nil safety check for price
	if pr.GetPrice() == nil {
		return nil, fmt.Errorf("inventory: product %s has no price", productID)
	}

	// Create Money value object from protobuf data
	price, err := valueobjects.NewMoney(pr.GetPrice().GetAmount(), pr.GetPrice().GetCurrency())
	if err != nil {
		return nil, fmt.Errorf("failed to create money from product price: %w", err)
	}

	return &productinfo.ProductInfo{
		Name:  pr.GetName(),
		Price: price,
	}, nil
}
