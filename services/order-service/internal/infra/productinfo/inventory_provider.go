package productinfoimpl

import (
	"context"
	"time"

	"order-service/internal/ports/productinfo"
	invpb "proto-go/inventory"
)

// InventoryProvider implements productinfo.Provider using inventory-service gRPC client.
type InventoryProvider struct {
	client invpb.InventoryServiceClient
}

func NewInventoryProvider(client invpb.InventoryServiceClient) productinfo.Provider {
	return &InventoryProvider{client: client}
}

func (p *InventoryProvider) GetProduct(ctx context.Context, productID string) (string, int64, string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	resp, err := p.client.GetProduct(ctx, &invpb.GetProductRequest{Id: productID})
	if err != nil || resp.GetProduct() == nil {
		return "", 0, "", err
	}
	pr := resp.GetProduct()
	return pr.GetName(), pr.GetPrice().GetAmount(), pr.GetPrice().GetCurrency(), nil
}
