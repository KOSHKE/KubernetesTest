package services

import (
	"context"

	"ecommerce-platform/pkg/common/errors"
	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/pkg/outbox"
	"ecommerce-platform/pkg/validation"
	"ecommerce-platform/services/inventory-service/internal/application/dto"
	"ecommerce-platform/services/inventory-service/internal/application/usecases"
	"ecommerce-platform/services/inventory-service/internal/domain/entities"
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
	validator     *validation.Validate
	logger        logger.Logger
}

// NewInventoryApplicationService creates a new inventory application service
func NewInventoryApplicationService(
	inventoryRepo repository.InventoryRepositoryFacade,
	logger logger.Logger,
) *InventoryApplicationService {
	// Create use cases
	createProductUseCase := usecases.NewCreateProductUseCase()
	addStockUseCase := usecases.NewAddStockUseCase()
	getProductUseCase := usecases.NewGetProductUseCase()
	reserveStockUseCase := usecases.NewReserveStockUseCase()
	releaseStockUseCase := usecases.NewReleaseStockUseCase()
	commitStockUseCase := usecases.NewCommitStockUseCase()

	v := validation.New()

	return &InventoryApplicationService{
		createProductUseCase: createProductUseCase,
		addStockUseCase:      addStockUseCase,
		getProductUseCase:    getProductUseCase,
		reserveStockUseCase:  reserveStockUseCase,
		releaseStockUseCase:  releaseStockUseCase,
		commitStockUseCase:   commitStockUseCase,
		inventoryRepo:        inventoryRepo,
		validator:            v,
		logger:               logger,
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
	stock, err := s.addStockUseCase.Execute(ctx, productID, quantity, s.inventoryRepo)
	if err != nil {
		s.logger.Error("failed to add stock", "productID", productID, "quantity", quantity, "error", err)
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
	// Validate request DTO
	if err := s.validator.Struct(req); err != nil {
		s.logger.Error("Invalid request parameters", "error", err)
		return nil, err
	}

	// Create product
	product, err := s.createProductUseCase.Execute(ctx, req.Name, req.Price, req.ImageURL, s.inventoryRepo)
	if err != nil {
		s.logger.Error("failed to create product", "name", req.Name, "error", err)
		return nil, err
	}

	// Add stock if specified
	var stockInfo dto.StockInfo
	if req.Stock != nil && req.Stock.AvailableQuantity > 0 {
		stock, err := s.addStockUseCase.Execute(ctx, product.ID, req.Stock.AvailableQuantity, s.inventoryRepo)
		if err != nil {
			return nil, err
		}
		stockInfo = dto.StockInfo{
			AvailableQuantity: stock.AvailableQuantity,
			ReservedQuantity:  stock.ReservedQuantity,
			TotalQuantity:     stock.AvailableQuantity + stock.ReservedQuantity,
		}
	}

	// Convert domain objects to DTO using the factory method
	return dto.NewProductResponse(product, stockInfo), nil
}

// GetProduct retrieves a product by ID
func (s *InventoryApplicationService) GetProduct(ctx context.Context, productID string) (*dto.ProductResponse, error) {

	product, err := s.getProductUseCase.Execute(ctx, productID, s.inventoryRepo)
	if err != nil {
		return nil, err
	}

	// Get stock information for this product
	stock, err := s.inventoryRepo.GetStockByProductID(ctx, productID)
	var stockInfo dto.StockInfo
	if err == nil {
		stockInfo = dto.StockInfo{
			AvailableQuantity: stock.AvailableQuantity,
			ReservedQuantity:  stock.ReservedQuantity,
			TotalQuantity:     stock.AvailableQuantity + stock.ReservedQuantity,
		}
	}

	// Convert entity to DTO using the factory method
	return dto.NewProductResponse(product, stockInfo), nil
}

// ReserveStock reserves stock for an order
func (s *InventoryApplicationService) ReserveStock(ctx context.Context, req *dto.ReserveStockRequest) (*dto.ReserveStockResponse, error) {
	// Validate request DTO
	if err := s.validator.Struct(req); err != nil {
		s.logger.Error("Invalid request parameters", "error", err)
		return nil, err
	}

	// Convert DTO to domain entities
	items := make([]valueobjects.Item, len(req.Items))
	for i, item := range req.Items {
		stockItem, err := valueobjects.NewItem(item.ProductID, item.Quantity)
		if err != nil {
			s.logger.Error("failed to create item", "productID", item.ProductID, "quantity", item.Quantity, "error", err)
			return nil, err
		}
		items[i] = *stockItem
	}

	// Use transaction to ensure both stock reservation and event are saved atomically
	err := s.inventoryRepo.WithTransaction(ctx, func(repo repository.InventoryRepositoryFacade) error {
		// Execute use case with transaction repository
		if err := s.reserveStockUseCase.Execute(ctx, req.OrderID, items, repo); err != nil {
			return err
		}

		// Save event to outbox table (to be published later)
		eventData := dto.StockEventDTO{
			OrderID: req.OrderID,
			Items:   items,
		}
		event := outbox.Event{
			AggregateID: req.OrderID,
			Type:        "StockReserved",
			Payload:     eventData,
		}

		// Create outbox service using transaction repository
		outboxService := outbox.NewService(repo, s.logger)
		if err := outboxService.SaveEvent(ctx, event); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		s.logger.Error("failed to reserve stock", "orderID", req.OrderID, "error", err)
		return nil, err
	}

	// Create success response
	reservedItems := make([]dto.StockReservationItem, len(items))
	for i, item := range items {
		reservedItems[i] = dto.StockReservationItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	response := &dto.ReserveStockResponse{
		OrderID:       req.OrderID,
		Success:       true,
		Message:       "Stock reserved successfully",
		ReservedItems: reservedItems,
		FailedItems:   []dto.StockReservationItem{},
	}

	return response, nil
}

// ListProducts retrieves a paginated list of products with stock information
// This method uses ProductInventory aggregate to solve N+1 problem
func (s *InventoryApplicationService) ListProducts(ctx context.Context, req *dto.ListProductsRequest) (*dto.ListProductsResponse, error) {
	// Validate request DTO
	if err := s.validator.Struct(req); err != nil {
		s.logger.Error("Invalid request parameters", "error", err)
		return nil, err
	}

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
				TotalQuantity:     productInventory.Stock.GetTotalQuantity(),
			}
		}

		productResponses[i] = *dto.NewProductResponse(productInventory.Product, stockInfo)
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
	// Validate request DTO
	if err := s.validator.Struct(req); err != nil {
		s.logger.Error("Invalid request parameters", "error", err)
		return nil, err
	}

	// Convert DTO to domain entities
	items := make([]valueobjects.Item, len(req.Items))
	for i, item := range req.Items {
		stockItem, err := valueobjects.NewItem(item.ProductID, item.Quantity)
		if err != nil {
			s.logger.Error("failed to create item", "productID", item.ProductID, "quantity", item.Quantity, "error", err)
			return nil, err
		}
		items[i] = *stockItem
	}

	// Use transaction to ensure both stock release and event are saved atomically
	err := s.inventoryRepo.WithTransaction(ctx, func(repo repository.InventoryRepositoryFacade) error {
		// Execute use case with transaction repository
		if err := s.releaseStockUseCase.Execute(ctx, req.OrderID, items, repo); err != nil {
			return err
		}

		// Save event to outbox table (to be published later)
		eventData := dto.StockEventDTO{
			OrderID: req.OrderID,
			Items:   items,
		}
		event := outbox.Event{
			AggregateID: req.OrderID,
			Type:        "StockReleased",
			Payload:     eventData,
		}

		// Create outbox service using transaction repository
		outboxService := outbox.NewService(repo, s.logger)
		if err := outboxService.SaveEvent(ctx, event); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		s.logger.Error("failed to release stock", "orderID", req.OrderID, "error", err)
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
	// Validate request DTO
	if err := s.validator.Struct(req); err != nil {
		s.logger.Error("Invalid request parameters", "error", err)
		return nil, err
	}

	// Convert DTO to domain entities
	items := make([]valueobjects.Item, len(req.Items))
	for i, item := range req.Items {
		stockItem, err := valueobjects.NewItem(item.ProductID, item.Quantity)
		if err != nil {
			s.logger.Error("failed to create item", "productID", item.ProductID, "quantity", item.Quantity, "error", err)
			return nil, err
		}
		items[i] = *stockItem
	}

	// Use transaction to ensure both stock commit and event are saved atomically
	err := s.inventoryRepo.WithTransaction(ctx, func(repo repository.InventoryRepositoryFacade) error {
		// Execute use case with transaction repository
		if err := s.commitStockUseCase.Execute(ctx, req.OrderID, items, repo); err != nil {
			return err
		}

		// Save event to outbox table (to be published later)
		eventData := dto.StockEventDTO{
			OrderID: req.OrderID,
			Items:   items,
		}
		event := outbox.Event{
			AggregateID: req.OrderID,
			Type:        "StockCommitted",
			Payload:     eventData,
		}

		// Create outbox service using transaction repository
		outboxService := outbox.NewService(repo, s.logger)
		if err := outboxService.SaveEvent(ctx, event); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		s.logger.Error("failed to commit stock", "orderID", req.OrderID, "error", err)
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
		s.logger.Error("failed to get stocks for availability check", "productIDs", productIDs, "error", err)
		return err
	}

	// Check availability for all items
	for _, item := range items {
		stock, exists := stocks[item.ProductID]
		if !exists {
			s.logger.Error("product not found during stock check", "productID", item.ProductID)
			return errors.ErrProductNotFound
		}
		if !stock.CanReserve(item.Quantity) {
			s.logger.Error("insufficient stock for availability check", "productID", item.ProductID, "requested", item.Quantity, "available", stock.AvailableQuantity)
			return errors.ErrInsufficientStock
		}
	}
	return nil
}
