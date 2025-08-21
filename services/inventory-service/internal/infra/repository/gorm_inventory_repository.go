package repository

import (
	"context"
	"strings"

	"inventory-service/internal/domain/models"

	"gorm.io/gorm"
)

type GormInventoryRepository struct{ db *gorm.DB }

func NewGormInventoryRepository(db *gorm.DB) *GormInventoryRepository {
	return &GormInventoryRepository{db: db}
}

func (r *GormInventoryRepository) GetProduct(ctx context.Context, id string) (*models.Product, error) {
	var p models.Product
	if err := r.db.WithContext(ctx).First(&p, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *GormInventoryRepository) ListProducts(ctx context.Context, categoryID string, page, limit int, search string) ([]*models.Product, int32, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	var products []*models.Product
	q := r.db.WithContext(ctx).Model(&models.Product{})
	if categoryID != "" {
		q = q.Where("category_id = ?", categoryID)
	}
	if s := strings.TrimSpace(search); s != "" {
		like := "%" + s + "%"
		q = q.Where("LOWER(name) LIKE LOWER(?) OR LOWER(description) LIKE LOWER(?)", like, like)
	}
	var total int64
	q.Count(&total)
	offset := (page - 1) * limit
	if err := q.Offset(offset).Limit(limit).Order("name ASC").Find(&products).Error; err != nil {
		return nil, 0, err
	}
	return products, int32(total), nil
}

func (r *GormInventoryRepository) GetStock(ctx context.Context, productID string) (*models.Stock, error) {
	var st models.Stock
	if err := r.db.WithContext(ctx).First(&st, "product_id = ?", productID).Error; err != nil {
		return nil, err
	}
	return &st, nil
}

func (r *GormInventoryRepository) GetStocksByIDs(ctx context.Context, productIDs []string) (map[string]*models.Stock, error) {
	result := make(map[string]*models.Stock, len(productIDs))
	if len(productIDs) == 0 {
		return result, nil
	}
	var stocks []models.Stock
	if err := r.db.WithContext(ctx).Where("product_id IN ?", productIDs).Find(&stocks).Error; err != nil {
		return nil, err
	}
	for i := range stocks {
		result[stocks[i].ProductID] = &stocks[i]
	}
	return result, nil
}

func (r *GormInventoryRepository) Reserve(ctx context.Context, productID string, qty int32) error {
	// Atomic update using a single SQL statement guarded by available quantity
	res := r.db.WithContext(ctx).Model(&models.Stock{}).
		Where("product_id = ? AND available_quantity >= ?", productID, qty).
		Updates(map[string]interface{}{
			"available_quantity": gorm.Expr("available_quantity - ?", qty),
			"reserved_quantity":  gorm.Expr("reserved_quantity + ?", qty),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GormInventoryRepository) Release(ctx context.Context, productID string, qty int32) error {
	res := r.db.WithContext(ctx).Model(&models.Stock{}).
		Where("product_id = ? AND reserved_quantity >= ?", productID, qty).
		Updates(map[string]interface{}{
			"available_quantity": gorm.Expr("available_quantity + ?", qty),
			"reserved_quantity":  gorm.Expr("reserved_quantity - ?", qty),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GormInventoryRepository) Commit(ctx context.Context, productID string, qty int32) error {
	res := r.db.WithContext(ctx).Model(&models.Stock{}).
		Where("product_id = ? AND reserved_quantity >= ?", productID, qty).
		UpdateColumn("reserved_quantity", gorm.Expr("reserved_quantity - ?", qty))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GormInventoryRepository) GetCategories(ctx context.Context, activeOnly bool) ([]*models.Category, error) {
	var categories []*models.Category
	q := r.db.WithContext(ctx).Model(&models.Category{})
	if activeOnly {
		q = q.Where("is_active = ?", true)
	}
	if err := q.Order("name ASC").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

// AutoMigrate creates tables
func (r *GormInventoryRepository) AutoMigrate() error {
	return r.db.AutoMigrate(&models.Product{}, &models.Stock{}, &models.Category{})
}
