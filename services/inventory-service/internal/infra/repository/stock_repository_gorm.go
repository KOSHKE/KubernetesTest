package repository

import (
	"context"

	"ecommerce-platform/services/inventory-service/internal/domain/entities"
	"ecommerce-platform/services/inventory-service/internal/infra/migration"

	"gorm.io/gorm"
)

// Use model from migration package

func recordFromStock(s *entities.Stock) migration.StockRecord {
	return migration.StockRecord{
		ProductID:         s.ProductID,
		AvailableQuantity: s.AvailableQuantity,
		ReservedQuantity:  s.ReservedQuantity,
		CreatedAt:         s.CreatedAt,
		UpdatedAt:         s.UpdatedAt,
	}
}

func stockFromRecord(r migration.StockRecord) *entities.Stock {
	return entities.NewStock(r.ProductID, r.AvailableQuantity, r.ReservedQuantity)
}

type GormStockRepository struct {
	db *gorm.DB
}

func NewGormStockRepository(db *gorm.DB) *GormStockRepository {
	return &GormStockRepository{db: db}
}

// WithTransaction executes operations within a database transaction

func (r *GormStockRepository) CreateStock(ctx context.Context, stock *entities.Stock) error {
	rec := recordFromStock(stock)
	result := r.db.WithContext(ctx).Create(&rec)
	return result.Error
}

func (r *GormStockRepository) GetStockByID(ctx context.Context, id string) (*entities.Stock, error) {
	var rec migration.StockRecord
	result := r.db.WithContext(ctx).First(&rec, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}
	return stockFromRecord(rec), nil
}

func (r *GormStockRepository) GetStockByProductID(ctx context.Context, productID string) (*entities.Stock, error) {
	var rec migration.StockRecord
	result := r.db.WithContext(ctx).First(&rec, "product_id = ?", productID)
	if result.Error != nil {
		return nil, result.Error
	}
	return stockFromRecord(rec), nil
}

func (r *GormStockRepository) UpdateStock(ctx context.Context, stock *entities.Stock) error {
	rec := recordFromStock(stock)
	result := r.db.WithContext(ctx).Save(&rec)
	return result.Error
}

func (r *GormStockRepository) GetByProductIDs(ctx context.Context, productIDs []string) (map[string]*entities.Stock, error) {
	var recs []migration.StockRecord
	result := r.db.WithContext(ctx).Where("product_id IN ?", productIDs).Find(&recs)
	if result.Error != nil {
		return nil, result.Error
	}

	stocks := make(map[string]*entities.Stock)
	for _, rec := range recs {
		stocks[rec.ProductID] = stockFromRecord(rec)
	}

	return stocks, nil
}

func (r *GormStockRepository) StockExistsByID(ctx context.Context, id string) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&migration.StockRecord{}).Where("id = ?", id).Count(&count)
	return count > 0, result.Error
}

func (r *GormStockRepository) ExistsByProductID(ctx context.Context, productID string) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&migration.StockRecord{}).Where("product_id = ?", productID).Count(&count)
	return count > 0, result.Error
}
