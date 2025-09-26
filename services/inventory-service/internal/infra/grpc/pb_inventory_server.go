package grpc

import (
	"context"
	"time"

	dto "ecommerce-platform/pkg/common/dto/inventory-service"
	"ecommerce-platform/pkg/common/grpcutils"
	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/proto-go/common"
	"ecommerce-platform/proto-go/inventory"
	"ecommerce-platform/services/inventory-service/internal/application/services"
	"ecommerce-platform/services/inventory-service/internal/metrics"
)

// PBInventoryServer implements the gRPC inventory service
type PBInventoryServer struct {
	inventory.UnimplementedInventoryServiceServer
	appService *services.InventoryApplicationService
	metrics    metrics.InventoryMetrics
}

// NewPBInventoryServer creates a new inventory gRPC server
func NewPBInventoryServer(appService *services.InventoryApplicationService, metrics metrics.InventoryMetrics) *PBInventoryServer {
	return &PBInventoryServer{
		appService: appService,
		metrics:    metrics,
	}
}

// GetProducts retrieves a paginated list of products
func (s *PBInventoryServer) GetProducts(ctx context.Context, req *inventory.GetProductsRequest) (*inventory.GetProductsResponse, error) {
	start := time.Now()
	method := "GetProducts"
	status := "success"
	defer func() {
		s.metrics.GRPCRequestDuration(method, time.Since(start))
		s.metrics.GRPCRequestTotal(method, status)
	}()

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
		status = "error"
		return nil, grpcutils.MapErrorToStatus(err)
	}

	// Convert response to gRPC
	grpcProducts := make([]*inventory.Product, len(response.Products))
	for i, product := range response.Products {
		grpcProducts[i] = &inventory.Product{
			Id:   product.ID,
			Name: product.Name,
			Price: &common.Money{
				Amount:   product.PriceAmount,
				Currency: product.Currency,
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
	start := time.Now()
	method := "GetProduct"
	status := "success"
	defer func() {
		s.metrics.GRPCRequestDuration(method, time.Since(start))
		s.metrics.GRPCRequestTotal(method, status)
	}()

	// Get product
	response, err := s.appService.GetProduct(ctx, req.Id)
	if err != nil {
		status = "error"
		return nil, grpcutils.MapErrorToStatus(err)
	}

	// Convert response to gRPC
	grpcResponse := &inventory.GetProductResponse{
		Product: &inventory.Product{
			Id:   response.ID,
			Name: response.Name,
			Price: &common.Money{
				Amount:   response.PriceAmount,
				Currency: response.Currency,
			},
			ImageUrl:      response.ImageURL,
			StockQuantity: response.Stock.AvailableQuantity,
		},
	}

	return grpcResponse, nil
}

// CheckStock checks stock availability for products
func (s *PBInventoryServer) CheckStock(ctx context.Context, req *inventory.CheckStockRequest) (*inventory.CheckStockResponse, error) {
	start := time.Now()
	method := "CheckStock"
	status := "success"
	defer func() {
		s.metrics.GRPCRequestDuration(method, time.Since(start))
		s.metrics.GRPCRequestTotal(method, status)
	}()

	// Convert gRPC request to domain value objects
	items := make([]valueobjects.Item, len(req.Items))
	for i, item := range req.Items {
		stockItem, err := valueobjects.NewItem(item.ProductId, item.Quantity)
		if err != nil {
			status = "error"
			return nil, grpcutils.MapErrorToStatus(err)
		}
		items[i] = *stockItem
	}

	// Get all product IDs for batch query
	productIDs := make([]string, len(items))
	for i, item := range items {
		productIDs[i] = item.ProductID
	}

	// Fetch all stocks in a single query
	stocks, err := s.appService.GetStocksByProductIDs(ctx, productIDs, false)
	if err != nil {
		status = "error"
		return nil, grpcutils.MapErrorToStatus(err)
	}
	allAvailable := true

	// Convert response to gRPC
	results := make([]*inventory.StockCheckResult, len(req.Items))

	for i, item := range items {
		stock, exists := stocks[item.ProductID]
		isAvailable := exists && stock.CanReserve(item.Quantity)

		if !isAvailable {
			allAvailable = false
		}

		// Get actual available quantity
		availableQuantity := int32(0)
		if exists {
			availableQuantity = stock.AvailableQuantity
		}

		results[i] = &inventory.StockCheckResult{
			ProductId:         item.ProductID,
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
	start := time.Now()
	method := "ReserveStock"
	status := "success"
	defer func() {
		s.metrics.GRPCRequestDuration(method, time.Since(start))
		s.metrics.GRPCRequestTotal(method, status)
	}()

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
		status = "error"
		return nil, grpcutils.MapErrorToStatus(err)
	}

	// Convert response to gRPC
	failedProducts := make([]string, len(response.FailedItems))
	for i, item := range response.FailedItems {
		failedProducts[i] = item.ProductID
	}

	grpcResponse := &inventory.ReserveStockResponse{
		Success:        response.Success,
		Message:        response.Message,
		FailedProducts: failedProducts,
	}

	return grpcResponse, nil
}

// ReleaseStock releases reserved stock
func (s *PBInventoryServer) ReleaseStock(ctx context.Context, req *inventory.ReleaseStockRequest) (*inventory.ReleaseStockResponse, error) {
	start := time.Now()
	method := "ReleaseStock"
	status := "success"
	defer func() {
		s.metrics.GRPCRequestDuration(method, time.Since(start))
		s.metrics.GRPCRequestTotal(method, status)
	}()

	// Convert gRPC request to domain value objects
	items := make([]valueobjects.Item, len(req.Items))
	for i, item := range req.Items {
		stockItem, err := valueobjects.NewItem(item.ProductId, item.Quantity)
		if err != nil {
			status = "error"
			return nil, grpcutils.MapErrorToStatus(err)
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
		status = "error"
		return nil, grpcutils.MapErrorToStatus(err)
	}

	grpcResponse := &inventory.ReleaseStockResponse{
		Success: true,
		Message: "Stock released successfully",
	}

	return grpcResponse, nil
}
