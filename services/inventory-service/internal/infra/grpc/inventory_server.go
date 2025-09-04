package grpc

import (
	"context"

	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/proto-go/inventory"
	"ecommerce-platform/services/inventory-service/internal/application/dto"
	"ecommerce-platform/services/inventory-service/internal/application/services"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// InventoryServer implements the gRPC inventory service
type InventoryServer struct {
	inventory.UnimplementedInventoryServiceServer
	appService *services.InventoryApplicationService
	logger     logger.Logger
}

// NewInventoryServer creates a new inventory gRPC server
func NewInventoryServer(appService *services.InventoryApplicationService, logger logger.Logger) *InventoryServer {
	return &InventoryServer{
		appService: appService,
		logger:     logger,
	}
}

// GetProducts retrieves a paginated list of products
func (s *InventoryServer) GetProducts(ctx context.Context, req *inventory.GetProductsRequest) (*inventory.GetProductsResponse, error) {
	s.logger.Info("getting products via gRPC", "page", req.Page, "limit", req.Limit)

	// Convert gRPC request to DTO
	listReq := &dto.ListProductsRequest{
		Page:   int(req.Page),
		Limit:  int(req.Limit),
		Search: req.Search,
	}

	// List products
	response, err := s.appService.ListProducts(ctx, listReq)
	if err != nil {
		s.logger.Error("failed to list products", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to list products: %v", err)
	}

	// Convert response to gRPC
	grpcProducts := make([]*inventory.Product, len(response.Products))
	for i, product := range response.Products {
		grpcProducts[i] = &inventory.Product{
			Id:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			Price: &inventory.Money{
				Amount:   product.Price.Amount,
				Currency: product.Price.Currency.String(),
			},
			ImageUrl:      product.ImageURL,
			IsActive:      product.IsActive,
			StockQuantity: product.Stock.AvailableQuantity,
		}
	}

	grpcResponse := &inventory.GetProductsResponse{
		Products: grpcProducts,
		Total:    response.Total,
	}

	s.logger.Info("products retrieved successfully via gRPC", "count", len(response.Products), "total", response.Total)
	return grpcResponse, nil
}

// GetProduct retrieves a product by ID
func (s *InventoryServer) GetProduct(ctx context.Context, req *inventory.GetProductRequest) (*inventory.GetProductResponse, error) {
	s.logger.Info("getting product via gRPC", "productID", req.Id)

	// Get product
	response, err := s.appService.GetProduct(ctx, req.Id)
	if err != nil {
		s.logger.Error("failed to get product", "productID", req.Id, "error", err)
		return nil, status.Errorf(codes.Internal, "failed to get product: %v", err)
	}

	// Convert response to gRPC
	grpcResponse := &inventory.GetProductResponse{
		Product: &inventory.Product{
			Id:          response.ID,
			Name:        response.Name,
			Description: response.Description,
			Price: &inventory.Money{
				Amount:   response.Price.Amount,
				Currency: response.Price.Currency.String(),
			},
			ImageUrl:      response.ImageURL,
			IsActive:      response.IsActive,
			StockQuantity: response.Stock.AvailableQuantity,
		},
	}

	s.logger.Info("product retrieved successfully via gRPC", "productID", req.Id)
	return grpcResponse, nil
}

// CheckStock checks stock availability for products
func (s *InventoryServer) CheckStock(ctx context.Context, req *inventory.CheckStockRequest) (*inventory.CheckStockResponse, error) {
	s.logger.Info("checking stock via gRPC", "itemsCount", len(req.Items))

	// Convert gRPC request to DTO
	items := make([]dto.StockReservationItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = dto.StockReservationItem{
			ProductID: item.ProductId,
			Quantity:  item.Quantity,
		}
	}

	// Check stock availability using domain service
	domainService := s.appService.GetDomainService()
	unavailableProducts, err := domainService.CheckStockAvailability(ctx, items)
	if err != nil {
		s.logger.Error("failed to check stock", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to check stock: %v", err)
	}

	// Convert response to gRPC
	results := make([]*inventory.StockCheckResult, len(req.Items))
	allAvailable := len(unavailableProducts) == 0

	for i, item := range req.Items {
		isAvailable := true
		for _, unavailable := range unavailableProducts {
			if unavailable == item.ProductId {
				isAvailable = false
				break
			}
		}

		results[i] = &inventory.StockCheckResult{
			ProductId:         item.ProductId,
			RequestedQuantity: item.Quantity,
			AvailableQuantity: 0, // TODO: Get actual available quantity
			IsAvailable:       isAvailable,
		}
	}

	grpcResponse := &inventory.CheckStockResponse{
		Results:      results,
		AllAvailable: allAvailable,
	}

	s.logger.Info("stock check completed via gRPC", "allAvailable", allAvailable)
	return grpcResponse, nil
}

// ReserveStock reserves stock for an order
func (s *InventoryServer) ReserveStock(ctx context.Context, req *inventory.ReserveStockRequest) (*inventory.ReserveStockResponse, error) {
	s.logger.Info("reserving stock via gRPC", "orderID", req.OrderId, "itemsCount", len(req.Items))

	// Convert gRPC request to DTO
	reserveReq := &dto.ReserveStockRequest{
		OrderID: req.OrderId,
		UserID:  req.UserId,
		Items:   make([]dto.StockReservationItem, len(req.Items)),
	}

	for i, item := range req.Items {
		reserveReq.Items[i] = dto.StockReservationItem{
			ProductID: item.ProductId,
			Quantity:  item.Quantity,
		}
	}

	// Reserve stock
	response, err := s.appService.ReserveStock(ctx, reserveReq)
	if err != nil {
		s.logger.Error("failed to reserve stock", "orderID", req.OrderId, "error", err)
		return nil, status.Errorf(codes.Internal, "failed to reserve stock: %v", err)
	}

	// Convert response to gRPC
	grpcResponse := &inventory.ReserveStockResponse{
		Success:        response.Success,
		Message:        response.Message,
		FailedProducts: response.FailedItems,
	}

	s.logger.Info("stock reservation completed via gRPC",
		"orderID", req.OrderId,
		"success", response.Success,
		"reservedCount", len(response.ReservedItems),
		"failedCount", len(response.FailedItems))

	return grpcResponse, nil
}

// ReleaseStock releases reserved stock
func (s *InventoryServer) ReleaseStock(ctx context.Context, req *inventory.ReleaseStockRequest) (*inventory.ReleaseStockResponse, error) {
	s.logger.Info("releasing stock via gRPC", "orderID", req.OrderId, "itemsCount", len(req.Items))

	// Convert gRPC request to DTO
	items := make([]dto.StockReservationItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = dto.StockReservationItem{
			ProductID: item.ProductId,
			Quantity:  item.Quantity,
		}
	}

	// Release stock using domain service
	domainService := s.appService.GetDomainService()
	err := domainService.ReleaseStock(ctx, req.OrderId, items)
	if err != nil {
		s.logger.Error("failed to release stock", "orderID", req.OrderId, "error", err)
		return nil, status.Errorf(codes.Internal, "failed to release stock: %v", err)
	}

	grpcResponse := &inventory.ReleaseStockResponse{
		Success: true,
		Message: "Stock released successfully",
	}

	s.logger.Info("stock released successfully via gRPC", "orderID", req.OrderId)
	return grpcResponse, nil
}
