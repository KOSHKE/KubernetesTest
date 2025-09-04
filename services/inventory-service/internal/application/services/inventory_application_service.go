package services

import (
	"context"
	"fmt"

	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/services/inventory-service/internal/application/dto"
	"ecommerce-platform/services/inventory-service/internal/application/usecases"
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

	// Domain service
	domainService *services.InventoryDomainService

	// Dependencies
	repo      repository.InventoryRepository
	publisher publisher.StockEventsPublisher
	logger    logger.Logger
}

// NewInventoryApplicationService creates a new inventory application service
func NewInventoryApplicationService(
	repo repository.InventoryRepository,
	publisher publisher.StockEventsPublisher,
	logger logger.Logger,
) *InventoryApplicationService {
	// Create domain service
	domainService := services.NewInventoryDomainService(repo, logger)

	// Create use cases
	createProductUseCase := usecases.NewCreateProductUseCase(repo, logger)
	getProductUseCase := usecases.NewGetProductUseCase(repo, logger)
	reserveStockUseCase := usecases.NewReserveStockUseCase(repo, publisher, domainService, logger)

	return &InventoryApplicationService{
		createProductUseCase: createProductUseCase,
		getProductUseCase:    getProductUseCase,
		reserveStockUseCase:  reserveStockUseCase,
		domainService:        domainService,
		repo:                 repo,
		publisher:            publisher,
		logger:               logger,
	}
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
	products, total, err := s.repo.ListProducts(ctx, req.Page, req.Limit, req.Search)
	if err != nil {
		s.logger.Error("failed to list products", "error", err)
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	// Convert to response DTOs
	productResponses := make([]dto.ProductResponse, len(products))
	for i, product := range products {
		productResponses[i] = dto.ProductResponse{
			ID:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
			ImageURL:    product.ImageURL,
			IsActive:    product.IsActive,
			Stock: dto.StockInfo{
				AvailableQuantity: product.GetAvailableQuantity(),
				ReservedQuantity:  product.GetReservedQuantity(),
				TotalQuantity:     product.GetTotalQuantity(),
			},
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

// GetDomainService returns the domain service for external use
func (s *InventoryApplicationService) GetDomainService() *services.InventoryDomainService {
	return s.domainService
}
