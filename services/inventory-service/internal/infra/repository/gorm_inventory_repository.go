package repository

import (
	"context"
	"fmt"

	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/services/inventory-service/internal/domain/aggregates"
	"ecommerce-platform/services/inventory-service/internal/domain/entities"
	"ecommerce-platform/services/inventory-service/internal/domain/ports/repository"
	"ecommerce-platform/services/inventory-service/internal/infra/migration"

	"gorm.io/gorm"
)

// GormInventoryRepository implements InventoryRepository using GORM
type GormInventoryRepository struct {
	db *gorm.DB
}

// NewGormInventoryRepository creates a new GORM inventory repository
func NewGormInventoryRepository(db *gorm.DB) repository.InventoryRepository {
	return &GormInventoryRepository{db: db}
}

// CreateProduct creates a new product
func (r *GormInventoryRepository) CreateProduct(ctx context.Context, product *aggregates.Product) error {
	// Convert aggregate to database record
	record := &migration.ProductRecord{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		PriceMinor:  product.Price.Amount,
		Currency:    product.Price.Currency.String(),
		ImageURL:    product.ImageURL,
		IsActive:    product.IsActive,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}

	// Create product
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	// Create stock record
	stockRecord := &migration.StockRecord{
		ProductID:         product.ID,
		AvailableQuantity: product.GetAvailableQuantity(),
		ReservedQuantity:  product.GetReservedQuantity(),
	}

	if err := r.db.WithContext(ctx).Create(stockRecord).Error; err != nil {
		return fmt.Errorf("failed to create stock: %w", err)
	}

	return nil
}

// GetProductByID retrieves a product by ID
func (r *GormInventoryRepository) GetProductByID(ctx context.Context, id string) (*aggregates.Product, error) {
	var productRecord migration.ProductRecord
	var stockRecord migration.StockRecord

	// Get product
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&productRecord).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Get stock
	if err := r.db.WithContext(ctx).Where("product_id = ?", id).First(&stockRecord).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create default stock if not found
			stockRecord = migration.StockRecord{
				ProductID:         id,
				AvailableQuantity: 0,
				ReservedQuantity:  0,
			}
		} else {
			return nil, fmt.Errorf("failed to get stock: %w", err)
		}
	}

	// Convert to aggregate
	product, err := r.recordToAggregate(&productRecord, &stockRecord)
	if err != nil {
		return nil, fmt.Errorf("failed to convert record to aggregate: %w", err)
	}

	return product, nil
}

// UpdateProduct updates a product
func (r *GormInventoryRepository) UpdateProduct(ctx context.Context, product *aggregates.Product) error {
	// Convert aggregate to database record
	record := &migration.ProductRecord{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		PriceMinor:  product.Price.Amount,
		Currency:    product.Price.Currency.String(),
		ImageURL:    product.ImageURL,
		IsActive:    product.IsActive,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}

	// Update product
	if err := r.db.WithContext(ctx).Save(record).Error; err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	// Update stock
	stockRecord := &migration.StockRecord{
		ProductID:         product.ID,
		AvailableQuantity: product.GetAvailableQuantity(),
		ReservedQuantity:  product.GetReservedQuantity(),
	}

	if err := r.db.WithContext(ctx).Save(stockRecord).Error; err != nil {
		return fmt.Errorf("failed to update stock: %w", err)
	}

	return nil
}

// DeleteProduct deletes a product
func (r *GormInventoryRepository) DeleteProduct(ctx context.Context, id string) error {
	// Delete stock first
	if err := r.db.WithContext(ctx).Where("product_id = ?", id).Delete(&migration.StockRecord{}).Error; err != nil {
		return fmt.Errorf("failed to delete stock: %w", err)
	}

	// Delete product
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&migration.ProductRecord{}).Error; err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}

// ListProducts retrieves a paginated list of products
func (r *GormInventoryRepository) ListProducts(ctx context.Context, page, limit int, search string) ([]*aggregates.Product, int32, error) {
	var productRecords []migration.ProductRecord
	var total int64

	query := r.db.WithContext(ctx).Model(&migration.ProductRecord{})

	// Apply filters
	if search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	// Apply pagination
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Find(&productRecords).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}

	// Convert to aggregates
	products := make([]*aggregates.Product, len(productRecords))
	for i, record := range productRecords {
		// Get stock for this product
		var stockRecord migration.StockRecord
		if err := r.db.WithContext(ctx).Where("product_id = ?", record.ID).First(&stockRecord).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Create default stock if not found
				stockRecord = migration.StockRecord{
					ProductID:         record.ID,
					AvailableQuantity: 0,
					ReservedQuantity:  0,
				}
			} else {
				return nil, 0, fmt.Errorf("failed to get stock for product %s: %w", record.ID, err)
			}
		}

		product, err := r.recordToAggregate(&record, &stockRecord)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to convert record to aggregate: %w", err)
		}
		products[i] = product
	}

	return products, int32(total), nil
}

// UpdateStock updates stock information
func (r *GormInventoryRepository) UpdateStock(ctx context.Context, productID string, availableQuantity, reservedQuantity int32) error {
	record := &migration.StockRecord{
		ProductID:         productID,
		AvailableQuantity: availableQuantity,
		ReservedQuantity:  reservedQuantity,
	}

	if err := r.db.WithContext(ctx).Save(record).Error; err != nil {
		return fmt.Errorf("failed to update stock: %w", err)
	}

	return nil
}

// GetStockByProductID retrieves stock information for a product
func (r *GormInventoryRepository) GetStockByProductID(ctx context.Context, productID string) (*aggregates.Product, error) {
	return r.GetProductByID(ctx, productID)
}

// GetStocksByProductIDs retrieves stock information for multiple products
func (r *GormInventoryRepository) GetStocksByProductIDs(ctx context.Context, productIDs []string) (map[string]*aggregates.Product, error) {
	if len(productIDs) == 0 {
		return make(map[string]*aggregates.Product), nil
	}

	var productRecords []migration.ProductRecord
	var stockRecords []migration.StockRecord

	// Get products
	if err := r.db.WithContext(ctx).Where("id IN ?", productIDs).Find(&productRecords).Error; err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}

	// Get stocks
	if err := r.db.WithContext(ctx).Where("product_id IN ?", productIDs).Find(&stockRecords).Error; err != nil {
		return nil, fmt.Errorf("failed to get stocks: %w", err)
	}

	// Create stock map
	stockMap := make(map[string]*migration.StockRecord)
	for _, stock := range stockRecords {
		stockMap[stock.ProductID] = &stock
	}

	// Convert to aggregates
	result := make(map[string]*aggregates.Product)
	for _, productRecord := range productRecords {
		stockRecord, exists := stockMap[productRecord.ID]
		if !exists {
			// Create default stock if not found
			stockRecord = &migration.StockRecord{
				ProductID:         productRecord.ID,
				AvailableQuantity: 0,
				ReservedQuantity:  0,
			}
		}

		product, err := r.recordToAggregate(&productRecord, stockRecord)
		if err != nil {
			return nil, fmt.Errorf("failed to convert record to aggregate: %w", err)
		}
		result[productRecord.ID] = product
	}

	return result, nil
}

// WithTx executes a function within a transaction
func (r *GormInventoryRepository) WithTx(ctx context.Context, fn func(repository.InventoryRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &GormInventoryRepository{db: tx}
		return fn(txRepo)
	})
}

// recordToAggregate converts database records to domain aggregate
func (r *GormInventoryRepository) recordToAggregate(productRecord *migration.ProductRecord, stockRecord *migration.StockRecord) (*aggregates.Product, error) {
	// Create money value object
	currency, err := valueobjects.ParseCurrency(productRecord.Currency)
	if err != nil {
		return nil, fmt.Errorf("failed to parse currency: %w", err)
	}
	money := valueobjects.NewMoney(productRecord.PriceMinor, currency)

	// Create product entity
	productEntity := entities.NewProduct(
		productRecord.ID,
		productRecord.Name,
		productRecord.Description,
		money,
		productRecord.ImageURL,
		productRecord.IsActive,
	)
	productEntity.CreatedAt = productRecord.CreatedAt
	productEntity.UpdatedAt = productRecord.UpdatedAt

	// Create stock value object
	stock := stockvalueobjects.NewStock(stockRecord.AvailableQuantity, stockRecord.ReservedQuantity)

	// Create aggregate
	product := &aggregates.Product{
		Product: productEntity,
		Stock:   stock,
	}

	return product, nil
}
