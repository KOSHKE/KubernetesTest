package services

import (
	"context"
	"fmt"

	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/services/inventory-service/internal/application/dto"
	"ecommerce-platform/services/inventory-service/internal/application/usecases"
	"ecommerce-platform/services/inventory-service/internal/domain/entities"
	"ecommerce-platform/services/inventory-service/internal/domain/ports/publisher"
	"ecommerce-platform/services/inventory-service/internal/domain/ports/repository"
	"ecommerce-platform/services/inventory-service/internal/domain/services"
)

// InventoryApplicationService orchestrates inventory operations
type InventoryApplicationService struct {
	// Use cases
	createProductUseCase *usecases.CreateProductUseCase
	getProductUseCase    *usecases.GetProductUseCase
	reserveStockUseCase  *usecases.ReserveStockUseCase
	releaseStockUseCase  *usecases.ReleaseStockUseCase
	commitStockUseCase   *usecases.CommitStockUseCase

	// Domain service
	domainService *services.InventoryDomainService

	// Dependencies
	inventoryRepo repository.InventoryRepository
	publisher     publisher.StockEventsPublisher
	logger        logger.Logger
}

// NewInventoryApplicationService creates a new inventory application service
func NewInventoryApplicationService(
	inventoryRepo repository.InventoryRepository,
	publisher publisher.StockEventsPublisher,
	logger logger.Logger,
) *InventoryApplicationService {
	// Create domain service
	domainService := services.NewInventoryDomainService(inventoryRepo, logger)

	// Create use cases
	createProductUseCase := usecases.NewCreateProductUseCase(inventoryRepo, logger)
	getProductUseCase := usecases.NewGetProductUseCase(inventoryRepo, logger)
	reserveStockUseCase := usecases.NewReserveStockUseCase(inventoryRepo, publisher, domainService, logger)
	releaseStockUseCase := usecases.NewReleaseStockUseCase(inventoryRepo, publisher, domainService, logger)
	commitStockUseCase := usecases.NewCommitStockUseCase(inventoryRepo, publisher, domainService, logger)

	return &InventoryApplicationService{
		createProductUseCase: createProductUseCase,
		getProductUseCase:    getProductUseCase,
		reserveStockUseCase:  reserveStockUseCase,
		releaseStockUseCase:  releaseStockUseCase,
		commitStockUseCase:   commitStockUseCase,
		domainService:        domainService,
		inventoryRepo:        inventoryRepo,
		publisher:            publisher,
		logger:               logger,
	}
}

// GetStockByProductID gets stock information for a product
func (s *InventoryApplicationService) GetStockByProductID(ctx context.Context, productID string) (*entities.Stock, error) {
	return s.inventoryRepo.GetStockByProductID(ctx, productID)
}

// CreateProduct creates a new product
func (s *InventoryApplicationService) CreateProduct(ctx context.Context, req *dto.CreateProductRequest) (*dto.ProductResponse, error) {
	s.logger.Info("creating product", "name", req.Name)

	response, err := s.createProductUseCase.Execute(ctx, req)
	if err != nil {
		s.logger.Error("failed to create product", "error", err)
		return nil, err
	}

	s.logger.Info("product created successfully", "productID", response.ID)
	return response, nil
}

// GetProduct retrieves a product by ID
func (s *InventoryApplicationService) GetProduct(ctx context.Context, productID string) (*dto.ProductResponse, error) {
	s.logger.Info("retrieving product", "productID", productID)

	response, err := s.getProductUseCase.Execute(ctx, productID)
	if err != nil {
		s.logger.Error("failed to get product", "productID", productID, "error", err)
		return nil, err
	}

	s.logger.Info("product retrieved successfully", "productID", productID)
	return response, nil
}

// ReserveStock reserves stock for an order
func (s *InventoryApplicationService) ReserveStock(ctx context.Context, req *dto.ReserveStockRequest) (*dto.ReserveStockResponse, error) {
	s.logger.Info("reserving stock", "orderID", req.OrderID, "itemsCount", len(req.Items))

	response, err := s.reserveStockUseCase.Execute(ctx, req)
	if err != nil {
		s.logger.Error("failed to reserve stock", "orderID", req.OrderID, "error", err)
		return nil, err
	}

	s.logger.Info("stock reservation completed",
		"orderID", req.OrderID,
		"success", response.Success,
		"reservedCount", len(response.ReservedItems),
		"failedCount", len(response.FailedItems))

	return response, nil
}

// ListProducts retrieves a paginated list of products
func (s *InventoryApplicationService) ListProducts(ctx context.Context, req *dto.ListProductsRequest) (*dto.ListProductsResponse, error) {
	s.logger.Info("listing products", "page", req.Page, "limit", req.Limit)

	// Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	// Get products from repository
	products, total, err := s.inventoryRepo.ListProducts(ctx, req.Page, req.Limit, req.Search)
	if err != nil {
		s.logger.Error("failed to list products", "error", err)
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	// Convert to response DTOs
	productResponses := make([]dto.ProductResponse, len(products))
	for i, product := range products {
		// Get stock information for this product
		stock, err := s.inventoryRepo.GetStockByProductID(ctx, product.ID)
		var stockInfo dto.StockInfo
		if err != nil {
			s.logger.Warn("failed to get stock info for product", "productID", product.ID, "error", err)
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

		productResponses[i] = dto.ProductResponse{
			ID:        product.ID,
			Name:      product.Name,
			Price:     product.Price,
			ImageURL:  product.ImageURL,
			Stock:     stockInfo,
			CreatedAt: product.CreatedAt,
			UpdatedAt: product.UpdatedAt,
		}
	}

	s.logger.Info("products listed successfully", "count", len(products), "total", total)

	return &dto.ListProductsResponse{
		Products: productResponses,
		Total:    total,
		Page:     req.Page,
		Limit:    req.Limit,
	}, nil
}

// ReleaseStock releases stock for an order
func (s *InventoryApplicationService) ReleaseStock(ctx context.Context, req *dto.ReleaseStockRequest) (*dto.ReleaseStockResponse, error) {
	s.logger.Info("releasing stock", "orderID", req.OrderID, "itemsCount", len(req.Items))

	response, err := s.releaseStockUseCase.Execute(ctx, req)
	if err != nil {
		s.logger.Error("failed to release stock", "orderID", req.OrderID, "error", err)
		return nil, err
	}

	s.logger.Info("stock released successfully", "orderID", req.OrderID)
	return response, nil
}

// CommitStock commits stock for an order
func (s *InventoryApplicationService) CommitStock(ctx context.Context, req *dto.CommitStockRequest) (*dto.CommitStockResponse, error) {
	s.logger.Info("committing stock", "orderID", req.OrderID, "itemsCount", len(req.Items))

	response, err := s.commitStockUseCase.Execute(ctx, req)
	if err != nil {
		s.logger.Error("failed to commit stock", "orderID", req.OrderID, "error", err)
		return nil, err
	}

	s.logger.Info("stock committed successfully", "orderID", req.OrderID)
	return response, nil
}

// CheckStockAvailability checks if products have sufficient stock
func (s *InventoryApplicationService) CheckStockAvailability(ctx context.Context, items []services.StockReservationItem) ([]string, error) {
	return s.domainService.CheckStockAvailability(ctx, items)
}
