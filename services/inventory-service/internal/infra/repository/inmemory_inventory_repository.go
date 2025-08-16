package repository

import (
	"context"
	"fmt"
	"strings"

	"inventory-service/internal/domain/models"
)

type InMemoryInventoryRepository struct {
	products map[string]*models.Product
	stocks   map[string]*models.Stock
}

func NewInMemoryInventoryRepository() *InMemoryInventoryRepository {
	return &InMemoryInventoryRepository{
		products: map[string]*models.Product{},
		stocks:   map[string]*models.Stock{},
	}
}

func (r *InMemoryInventoryRepository) Seed(p []*models.Product, s []*models.Stock) {
	for _, pr := range p {
		r.products[pr.ID] = pr
	}
	for _, st := range s {
		r.stocks[st.ProductID] = st
	}
}

func (r *InMemoryInventoryRepository) GetProduct(ctx context.Context, id string) (*models.Product, error) {
	if v, ok := r.products[id]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("product not found")
}

func (r *InMemoryInventoryRepository) ListProducts(ctx context.Context, categoryID string, page, limit int, search string) ([]*models.Product, int32, error) {
	var list []*models.Product
	for _, pr := range r.products {
		if categoryID != "" && pr.CategoryID != categoryID {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(pr.Name), strings.ToLower(search)) {
			continue
		}
		list = append(list, pr)
	}
	total := int32(len(list))
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	start := (page - 1) * limit
	if start > len(list) {
		return []*models.Product{}, total, nil
	}
	end := start + limit
	if end > len(list) {
		end = len(list)
	}
	return list[start:end], total, nil
}

func (r *InMemoryInventoryRepository) GetStock(ctx context.Context, productID string) (*models.Stock, error) {
	if v, ok := r.stocks[productID]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("stock not found")
}

func (r *InMemoryInventoryRepository) Reserve(ctx context.Context, productID string, qty int32) error {
	st, ok := r.stocks[productID]
	if !ok {
		return fmt.Errorf("stock not found")
	}
	return st.Reserve(qty)
}

func (r *InMemoryInventoryRepository) Release(ctx context.Context, productID string, qty int32) error {
	st, ok := r.stocks[productID]
	if !ok {
		return fmt.Errorf("stock not found")
	}
	return st.Release(qty)
}
