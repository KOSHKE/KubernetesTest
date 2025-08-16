package grpc

import (
	"context"
	appsvc "inventory-service/internal/app/services"
	"inventory-service/internal/domain/models"
	invpb "proto-go/inventory"
)

type PBInventoryServer struct {
	invpb.UnimplementedInventoryServiceServer
	svc *appsvc.InventoryService
}

func NewPBInventoryServer(svc *appsvc.InventoryService) *PBInventoryServer {
	return &PBInventoryServer{svc: svc}
}

func (s *PBInventoryServer) GetProducts(ctx context.Context, req *invpb.GetProductsRequest) (*invpb.GetProductsResponse, error) {
	products, total, err := s.svc.ListProducts(ctx, req.CategoryId, int(req.Page), int(req.Limit), req.Search)
	if err != nil {
		return nil, err
	}
	out := make([]*invpb.Product, 0, len(products))
	for _, p := range products {
		out = append(out, mapProductToPB(p))
	}
	return &invpb.GetProductsResponse{Products: out, Total: int32(total)}, nil
}

func (s *PBInventoryServer) GetProduct(ctx context.Context, req *invpb.GetProductRequest) (*invpb.GetProductResponse, error) {
	p, err := s.svc.GetProduct(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &invpb.GetProductResponse{Product: mapProductToPB(p)}, nil
}

func (s *PBInventoryServer) CheckStock(ctx context.Context, req *invpb.CheckStockRequest) (*invpb.CheckStockResponse, error) {
	items := make([]appsvc.StockCheckItem, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, appsvc.StockCheckItem{ProductID: it.ProductId, Quantity: it.Quantity})
	}
	results, all, err := s.svc.CheckStock(ctx, items)
	if err != nil {
		return nil, err
	}
	out := make([]*invpb.StockCheckResult, 0, len(results))
	for _, r := range results {
		out = append(out, &invpb.StockCheckResult{ProductId: r.ProductID, RequestedQuantity: r.RequestedQuantity, AvailableQuantity: r.AvailableQuantity, IsAvailable: r.IsAvailable})
	}
	return &invpb.CheckStockResponse{Results: out, AllAvailable: all}, nil
}

func (s *PBInventoryServer) ReserveStock(ctx context.Context, req *invpb.ReserveStockRequest) (*invpb.ReserveStockResponse, error) {
	items := make([]appsvc.StockCheckItem, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, appsvc.StockCheckItem{ProductID: it.ProductId, Quantity: it.Quantity})
	}
	failed, err := s.svc.ReserveStock(ctx, req.OrderId, items)
	if err != nil {
		return nil, err
	}
	ok := len(failed) == 0
	msg := "Reservation successful"
	if !ok {
		msg = "Reservation partial failure"
	}
	return &invpb.ReserveStockResponse{Success: ok, Message: msg, FailedProducts: failed}, nil
}

func (s *PBInventoryServer) ReleaseStock(ctx context.Context, req *invpb.ReleaseStockRequest) (*invpb.ReleaseStockResponse, error) {
	items := make([]appsvc.StockCheckItem, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, appsvc.StockCheckItem{ProductID: it.ProductId, Quantity: it.Quantity})
	}
	if err := s.svc.ReleaseStock(ctx, req.OrderId, items); err != nil {
		return nil, err
	}
	return &invpb.ReleaseStockResponse{Success: true, Message: "Released"}, nil
}

// mapping helpers
func mapProductToPB(p *models.Product) *invpb.Product {
	return &invpb.Product{
		Id:            p.ID,
		Name:          p.Name,
		Description:   p.Description,
		Price:         &invpb.Money{Amount: p.PriceMinor, Currency: p.Currency},
		CategoryId:    p.CategoryID,
		CategoryName:  p.CategoryName,
		StockQuantity: 0,
		ImageUrl:      p.ImageURL,
		IsActive:      p.IsActive,
	}
}
