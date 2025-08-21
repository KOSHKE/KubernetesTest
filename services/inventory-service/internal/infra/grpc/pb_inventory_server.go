package grpc

import (
	"context"
	"errors"
	appsvc "inventory-service/internal/app/services"
	"inventory-service/internal/domain/models"
	invpb "proto-go/inventory"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type PBInventoryServer struct {
	invpb.UnimplementedInventoryServiceServer
	svc *appsvc.InventoryService
	log *zap.SugaredLogger
}

func NewPBInventoryServer(svc *appsvc.InventoryService, log *zap.SugaredLogger) *PBInventoryServer {
	return &PBInventoryServer{svc: svc, log: log}
}

func (s *PBInventoryServer) GetProducts(ctx context.Context, req *invpb.GetProductsRequest) (*invpb.GetProductsResponse, error) {
	page := int(req.Page)
	if page <= 0 {
		page = 1
	}
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 20
	}
	products, total, err := s.svc.ListProducts(ctx, req.CategoryId, page, limit, req.Search)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list products: %v", err)
	}
	out := make([]*invpb.Product, 0, len(products))
	ids := make([]string, 0, len(products))
	for _, p := range products {
		ids = append(ids, p.ID)
	}
	stockMap := map[string]*models.Stock{}
	if len(ids) > 0 {
		if stocks, e := s.svc.GetStocksByIDs(ctx, ids); e == nil {
			stockMap = stocks
		} else {
			s.log.Warnw("failed to get stocks", "error", e)
		}
	}
	for _, p := range products {
		if st, ok := stockMap[p.ID]; ok {
			out = append(out, mapProductToPB(p, st.AvailableQuantity))
			continue
		}
		// degrade to 0 if no stock or error during batch fetch
		out = append(out, mapProductToPB(p, 0))
	}
	return &invpb.GetProductsResponse{Products: out, Total: int32(total)}, nil
}

func (s *PBInventoryServer) GetProduct(ctx context.Context, req *invpb.GetProductRequest) (*invpb.GetProductResponse, error) {
	p, err := s.svc.GetProduct(ctx, req.Id)
	if err != nil {
		code := codes.Internal
		if errors.Is(err, gorm.ErrRecordNotFound) {
			code = codes.NotFound
		}
		return nil, status.Errorf(code, "failed to get product: %v", err)
	}
	q, err := s.svc.GetStockQuantity(ctx, p.ID)
	if err != nil {
		return &invpb.GetProductResponse{Product: mapProductToPB(p, 0)}, nil
	}
	return &invpb.GetProductResponse{Product: mapProductToPB(p, q)}, nil
}

func (s *PBInventoryServer) CheckStock(ctx context.Context, req *invpb.CheckStockRequest) (*invpb.CheckStockResponse, error) {
	items := make([]appsvc.StockCheckItem, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, appsvc.StockCheckItem{ProductID: it.ProductId, Quantity: it.Quantity})
	}
	results, all, err := s.svc.CheckStock(ctx, items)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check stock: %v", err)
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
	failed, err := s.svc.ReserveStock(ctx, req.OrderId, req.UserId, items)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to reserve stock: %v", err)
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
		return nil, status.Errorf(codes.Internal, "failed to release stock: %v", err)
	}
	return &invpb.ReleaseStockResponse{Success: true, Message: "Released"}, nil
}

// mapping helpers
func mapProductToPB(p *models.Product, stockQty int32) *invpb.Product {
	return &invpb.Product{
		Id:            p.ID,
		Name:          p.Name,
		Description:   p.Description,
		Price:         &invpb.Money{Amount: p.PriceMinor, Currency: p.Currency},
		CategoryId:    p.CategoryID,
		CategoryName:  p.CategoryName,
		StockQuantity: stockQty,
		ImageUrl:      p.ImageURL,
		IsActive:      p.IsActive,
	}
}

func (s *PBInventoryServer) GetCategories(ctx context.Context, req *invpb.GetCategoriesRequest) (*invpb.GetCategoriesResponse, error) {
	categories, err := s.svc.GetCategories(ctx, req.ActiveOnly)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get categories: %v", err)
	}
	
	out := make([]*invpb.Category, 0, len(categories))
	for _, c := range categories {
		out = append(out, &invpb.Category{
			Id:          c.ID,
			Name:        c.Name,
			Description: c.Description,
			IsActive:    c.IsActive,
		})
	}
	
	return &invpb.GetCategoriesResponse{Categories: out}, nil
}
