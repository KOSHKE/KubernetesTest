package grpc

import (
	"context"

	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/proto-go/inventory"
	"ecommerce-platform/services/inventory-service/internal/application/dto"
	"ecommerce-platform/services/inventory-service/internal/application/services"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PBInventoryServer implements the gRPC inventory service
type PBInventoryServer struct {
	inventory.UnimplementedInventoryServiceServer
	appService *services.InventoryApplicationService
	logger     logger.Logger
}

// NewPBInventoryServer creates a new inventory gRPC server
func NewPBInventoryServer(appService *services.InventoryApplicationService, logger logger.Logger) *PBInventoryServer {
	return &PBInventoryServer{
		appService: appService,
		logger:     logger,
	}
}

// GetProducts retrieves a paginated list of products
func (s *PBInventoryServer) GetProducts(ctx context.Context, req *inventory.GetProductsRequest) (*inventory.GetProductsResponse, error) {
	s.logger.Info("getting products via gRPC", "page", req.Page, "limit", req.Limit)

	// Validate and set defaults
	page := int(req.Page)
	limit := int(req.Limit)

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	// Convert gRPC request to DTO
	listReq := &dto.ListProductsRequest{
		Page:   page,
		Limit:  limit,
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
			Id:   product.ID,
			Name: product.Name,
			Price: &inventory.Money{
				Amount:   product.Price.Amount,
				Currency: product.Price.Currency.String(),
			},
			ImageUrl:      product.ImageURL,
			StockQuantity: product.Stock.AvailableQuantity,
		}
	}

	grpcResponse := &inventory.GetProductsResponse{
		Products: grpcProducts,
		Total:    response.Total,
	}

	return grpcResponse, nil
}

// GetProduct retrieves a product by ID
func (s *PBInventoryServer) GetProduct(ctx context.Context, req *inventory.GetProductRequest) (*inventory.GetProductResponse, error) {
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
			Id:   response.ID,
			Name: response.Name,
			Price: &inventory.Money{
				Amount:   response.Price.Amount,
				Currency: response.Price.Currency.String(),
			},
			ImageUrl:      response.ImageURL,
			StockQuantity: response.Stock.AvailableQuantity,
		},
	}

	return grpcResponse, nil
}

// CheckStock checks stock availability for products
func (s *PBInventoryServer) CheckStock(ctx context.Context, req *inventory.CheckStockRequest) (*inventory.CheckStockResponse, error) {
	s.logger.Info("checking stock via gRPC", "itemsCount", len(req.Items))

	// Convert gRPC request to domain value objects
	items := make([]valueobjects.Item, len(req.Items))
	for i, item := range req.Items {
		stockItem, err := valueobjects.NewItem(item.ProductId, item.Quantity)
		if err != nil {
			s.logger.Error("invalid item data", "productID", item.ProductId, "quantity", item.Quantity, "error", err)
			return nil, status.Errorf(codes.InvalidArgument, "invalid item data: %v", err)
		}
		items[i] = *stockItem
	}

	// Check stock availability using application service
	err := s.appService.CheckStockAvailability(ctx, items)
	allAvailable := err == nil

	// Convert response to gRPC
	results := make([]*inventory.StockCheckResult, len(req.Items))

	for i, item := range req.Items {
		isAvailable := true
		if err != nil {
			// If there's an error, check individual items
			stock, stockErr := s.appService.GetStockByProductID(ctx, item.ProductId)
			if stockErr != nil || !stock.CanReserve(item.Quantity) {
				isAvailable = false
			}
		}

		// Get actual available quantity
		availableQuantity := int32(0)
		if isAvailable {
			stock, err := s.appService.GetStockByProductID(ctx, item.ProductId)
			if err == nil && stock != nil {
				availableQuantity = stock.AvailableQuantity
			}
		}

		results[i] = &inventory.StockCheckResult{
			ProductId:         item.ProductId,
			RequestedQuantity: item.Quantity,
			AvailableQuantity: availableQuantity,
			IsAvailable:       isAvailable,
		}
	}

	grpcResponse := &inventory.CheckStockResponse{
		Results:      results,
		AllAvailable: allAvailable,
	}

	return grpcResponse, nil
}

// ReserveStock reserves stock for an order
func (s *PBInventoryServer) ReserveStock(ctx context.Context, req *inventory.ReserveStockRequest) (*inventory.ReserveStockResponse, error) {
	s.logger.Info("reserving stock via gRPC", "orderID", req.OrderId, "itemsCount", len(req.Items))

	// Convert gRPC request to DTO
	reserveReq := &dto.ReserveStockRequest{
		OrderID: req.OrderId,
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

	return grpcResponse, nil
}

// ReleaseStock releases reserved stock
func (s *PBInventoryServer) ReleaseStock(ctx context.Context, req *inventory.ReleaseStockRequest) (*inventory.ReleaseStockResponse, error) {
	s.logger.Info("releasing stock via gRPC", "orderID", req.OrderId, "itemsCount", len(req.Items))

	// Convert gRPC request to domain value objects
	items := make([]valueobjects.Item, len(req.Items))
	for i, item := range req.Items {
		stockItem, err := valueobjects.NewItem(item.ProductId, item.Quantity)
		if err != nil {
			s.logger.Error("invalid item data", "productID", item.ProductId, "quantity", item.Quantity, "error", err)
			return nil, status.Errorf(codes.InvalidArgument, "invalid item data: %v", err)
		}
		items[i] = *stockItem
	}

	// Convert to DTO and release stock using application service
	releaseReq := &dto.ReleaseStockRequest{
		OrderID: req.OrderId,
		Items:   make([]dto.StockReservationItem, len(items)),
	}
	for i, item := range items {
		releaseReq.Items[i] = dto.StockReservationItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	_, err := s.appService.ReleaseStock(ctx, releaseReq)
	if err != nil {
		s.logger.Error("failed to release stock", "orderID", req.OrderId, "error", err)
		return nil, status.Errorf(codes.Internal, "failed to release stock: %v", err)
	}

	grpcResponse := &inventory.ReleaseStockResponse{
		Success: true,
		Message: "Stock released successfully",
	}

	return grpcResponse, nil
}
