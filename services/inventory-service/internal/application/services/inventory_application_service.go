package services

import (
	"context"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/pkg/outbox"
	"ecommerce-platform/services/inventory-service/internal/application/dto"
	"ecommerce-platform/services/inventory-service/internal/application/usecases"
	"ecommerce-platform/services/inventory-service/internal/domain/entities"
	"ecommerce-platform/services/inventory-service/internal/domain/ports/publisher"
	"ecommerce-platform/services/inventory-service/internal/domain/ports/repository"
)

// InventoryApplicationService orchestrates inventory operations
type InventoryApplicationService struct {
	// Use cases
	createProductUseCase *usecases.CreateProductUseCase
	addStockUseCase      *usecases.AddStockUseCase
	getProductUseCase    *usecases.GetProductUseCase
	reserveStockUseCase  *usecases.ReserveStockUseCase
	releaseStockUseCase  *usecases.ReleaseStockUseCase
	commitStockUseCase   *usecases.CommitStockUseCase

	// Dependencies
	inventoryRepo repository.InventoryRepositoryFacade
	publisher     publisher.StockEventsPublisher
}

// NewInventoryApplicationService creates a new inventory application service
func NewInventoryApplicationService(
	inventoryRepo repository.InventoryRepositoryFacade,
	publisher publisher.StockEventsPublisher,
	logger logger.Logger,
) *InventoryApplicationService {
	// Create outbox service using inventory repo (which includes outbox)
	outboxService := outbox.NewService(inventoryRepo, logger)

	// Create use cases
	createProductUseCase := usecases.NewCreateProductUseCase(inventoryRepo, logger)
	addStockUseCase := usecases.NewAddStockUseCase(inventoryRepo, logger)
	getProductUseCase := usecases.NewGetProductUseCase(inventoryRepo, logger)
	reserveStockUseCase := usecases.NewReserveStockUseCase(inventoryRepo, outboxService, logger)
	releaseStockUseCase := usecases.NewReleaseStockUseCase(inventoryRepo, outboxService, logger)
	commitStockUseCase := usecases.NewCommitStockUseCase(inventoryRepo, outboxService, logger)

	return &InventoryApplicationService{
		createProductUseCase: createProductUseCase,
		addStockUseCase:      addStockUseCase,
		getProductUseCase:    getProductUseCase,
		reserveStockUseCase:  reserveStockUseCase,
		releaseStockUseCase:  releaseStockUseCase,
		commitStockUseCase:   commitStockUseCase,
		inventoryRepo:        inventoryRepo,
		publisher:            publisher,
	}
}

// GetStockByProductID gets stock information for a product
func (s *InventoryApplicationService) GetStockByProductID(ctx context.Context, productID string) (*entities.Stock, error) {
	return s.inventoryRepo.GetStockByProductID(ctx, productID)
}

// GetStocksByProductIDs gets stock information for multiple products
func (s *InventoryApplicationService) GetStocksByProductIDs(ctx context.Context, productIDs []string, forUpdate bool) (map[string]*entities.Stock, error) {
	return s.inventoryRepo.GetStocksByProductIDs(ctx, productIDs, forUpdate)
}

// AddStock adds stock to a product
func (s *InventoryApplicationService) AddStock(ctx context.Context, productID string, quantity int32) (*dto.StockInfo, error) {

	stock, err := s.addStockUseCase.Execute(ctx, productID, quantity)
	if err != nil {
		return nil, err
	}

	stockInfo := dto.StockInfo{
		AvailableQuantity: stock.AvailableQuantity,
		ReservedQuantity:  stock.ReservedQuantity,
		TotalQuantity:     stock.AvailableQuantity + stock.ReservedQuantity,
	}

	return &stockInfo, nil
}

// CreateProduct creates a new product
func (s *InventoryApplicationService) CreateProduct(ctx context.Context, req *dto.CreateProductRequest) (*dto.ProductResponse, error) {

	// Create product
	product, err := s.createProductUseCase.Execute(ctx, req.Name, req.Price, req.ImageURL)
	if err != nil {
		return nil, err
	}

	// Add stock if specified
	var stockInfo dto.StockInfo
	if req.Stock > 0 {
		stock, err := s.addStockUseCase.Execute(ctx, product.ID, req.Stock)
		if err != nil {
			return nil, err
		}
		stockInfo = dto.StockInfo{
			AvailableQuantity: stock.AvailableQuantity,
			ReservedQuantity:  stock.ReservedQuantity,
			TotalQuantity:     stock.AvailableQuantity + stock.ReservedQuantity,
		}
	} else {
		stockInfo = dto.StockInfo{
			AvailableQuantity: 0,
			ReservedQuantity:  0,
			TotalQuantity:     0,
		}
	}

	// Convert domain objects to DTO
	response := &dto.ProductResponse{
		ID:        product.ID,
		Name:      product.Name,
		Price:     product.Price,
		ImageURL:  product.ImageURL,
		Stock:     stockInfo,
		CreatedAt: product.CreatedAt,
		UpdatedAt: product.UpdatedAt,
	}

	return response, nil
}

// GetProduct retrieves a product by ID
func (s *InventoryApplicationService) GetProduct(ctx context.Context, productID string) (*dto.ProductResponse, error) {

	product, err := s.getProductUseCase.Execute(ctx, productID)
	if err != nil {
		return nil, err
	}

	// Get stock information for this product
	stock, err := s.inventoryRepo.GetStockByProductID(ctx, productID)
	var stockInfo dto.StockInfo
	if err != nil {
		stockInfo = dto.StockInfo{
			AvailableQuantity: 0,
			ReservedQuantity:  0,
			TotalQuantity:     0,
		}
	} else {
		stockInfo = dto.StockInfo{
			AvailableQuantity: stock.AvailableQuantity,
			ReservedQuantity:  stock.ReservedQuantity,
			TotalQuantity:     stock.AvailableQuantity + stock.ReservedQuantity,
		}
	}

	// Convert entity to DTO
	response := &dto.ProductResponse{
		ID:        product.ID,
		Name:      product.Name,
		Price:     product.Price,
		ImageURL:  product.ImageURL,
		Stock:     stockInfo,
		CreatedAt: product.CreatedAt,
		UpdatedAt: product.UpdatedAt,
	}

	return response, nil
}

// ReserveStock reserves stock for an order
func (s *InventoryApplicationService) ReserveStock(ctx context.Context, req *dto.ReserveStockRequest) (*dto.ReserveStockResponse, error) {
	// Convert DTO to domain entities
	items := make([]valueobjects.Item, len(req.Items))
	for i, item := range req.Items {
		stockItem, err := valueobjects.NewItem(item.ProductID, item.Quantity)
		if err != nil {
			return nil, err
		}
		items[i] = *stockItem
	}

	err := s.reserveStockUseCase.Execute(ctx, req.OrderID, items)
	if err != nil {
		return nil, err
	}

	// Create success response
	response := &dto.ReserveStockResponse{
		OrderID:       req.OrderID,
		Success:       true,
		Message:       "Stock reserved successfully",
		ReservedItems: make([]string, len(items)),
		FailedItems:   []string{},
	}

	// Add all product IDs as reserved
	for i, item := range items {
		response.ReservedItems[i] = item.ProductID
	}

	return response, nil
}

// ListProducts retrieves a paginated list of products with stock information
// This method uses ProductInventory aggregate to solve N+1 problem
func (s *InventoryApplicationService) ListProducts(ctx context.Context, req *dto.ListProductsRequest) (*dto.ListProductsResponse, error) {
	// Get products with stock information using single query
	productInventories, total, err := s.inventoryRepo.ListProductsWithStock(ctx, req.Page, req.Limit, req.Search)
	if err != nil {
		return nil, err
	}

	// Convert to response DTOs
	productResponses := make([]dto.ProductResponse, len(productInventories))
	for i, productInventory := range productInventories {
		// Handle stock information safely
		var stockInfo dto.StockInfo
		if productInventory.Stock != nil {
			stockInfo = dto.StockInfo{
				AvailableQuantity: productInventory.Stock.AvailableQuantity,
				ReservedQuantity:  productInventory.Stock.ReservedQuantity,
				TotalQuantity:     productInventory.GetTotalQuantity(),
			}
		}

		productResponses[i] = dto.ProductResponse{
			ID:        productInventory.Product.ID,
			Name:      productInventory.Product.Name,
			Price:     productInventory.Product.Price,
			ImageURL:  productInventory.Product.ImageURL,
			Stock:     stockInfo,
			CreatedAt: productInventory.Product.CreatedAt,
			UpdatedAt: productInventory.Product.UpdatedAt,
		}
	}

	return &dto.ListProductsResponse{
		Products: productResponses,
		Total:    total,
		Page:     req.Page,
		Limit:    req.Limit,
	}, nil
}

// ReleaseStock releases stock for an order
func (s *InventoryApplicationService) ReleaseStock(ctx context.Context, req *dto.ReleaseStockRequest) (*dto.ReleaseStockResponse, error) {

	// Convert DTO to domain entities
	items := make([]valueobjects.Item, len(req.Items))
	for i, item := range req.Items {
		stockItem, err := valueobjects.NewItem(item.ProductID, item.Quantity)
		if err != nil {
			return nil, err
		}
		items[i] = *stockItem
	}

	err := s.releaseStockUseCase.Execute(ctx, req.OrderID, items)
	if err != nil {
		return nil, err
	}

	return &dto.ReleaseStockResponse{
		OrderID: req.OrderID,
		Success: true,
		Message: "Stock released successfully",
	}, nil
}

// CommitStock commits stock for an order
func (s *InventoryApplicationService) CommitStock(ctx context.Context, req *dto.CommitStockRequest) (*dto.CommitStockResponse, error) {

	// Convert DTO to domain entities
	items := make([]valueobjects.Item, len(req.Items))
	for i, item := range req.Items {
		stockItem, err := valueobjects.NewItem(item.ProductID, item.Quantity)
		if err != nil {
			return nil, err
		}
		items[i] = *stockItem
	}

	err := s.commitStockUseCase.Execute(ctx, req.OrderID, items)
	if err != nil {
		return nil, err
	}

	return &dto.CommitStockResponse{
		OrderID: req.OrderID,
		Success: true,
		Message: "Stock committed successfully",
	}, nil
}

// CheckStockAvailability checks if products have sufficient stock
func (s *InventoryApplicationService) CheckStockAvailability(ctx context.Context, items []valueobjects.Item) error {
	// Get all product IDs for batch query
	productIDs := make([]string, len(items))
	for i, item := range items {
		productIDs[i] = item.ProductID
	}

	// Fetch all stocks in a single query
	stocks, err := s.inventoryRepo.GetStocksByProductIDs(ctx, productIDs, false)
	if err != nil {
		return err
	}

	// Check availability for all items
	for _, item := range items {
		stock, exists := stocks[item.ProductID]
		if !exists {
			return errors.ErrProductNotFound
		}
		if !stock.CanReserve(item.Quantity) {
			return errors.ErrInsufficientStock
		}
	}
	return nil
}
