package repository

import (
	"context"

	"ecommerce-platform/pkg/common/valueobjects"
	"ecommerce-platform/services/inventory-service/internal/domain/entities"
	"ecommerce-platform/services/inventory-service/internal/infra/migration"

	"gorm.io/gorm"
)

// Use model from migration package

func recordFromProduct(p *entities.Product) migration.ProductRecord {
	return migration.ProductRecord{
		ID:        p.ID,
		Name:      p.Name,
		Price:     p.Price.Amount,
		Currency:  p.Price.Currency.Code,
		ImageURL:  p.ImageURL,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func productFromRecord(r migration.ProductRecord) (*entities.Product, error) {
	currency, err := valueobjects.NewCurrency(r.Currency)
	if err != nil {
		return nil, err
	}
	price := valueobjects.NewMoney(r.Price, currency)
	return entities.NewProduct(r.ID, r.Name, price, r.ImageURL), nil
}

type GormProductRepository struct {
	db *gorm.DB
}

func NewGormProductRepository(db *gorm.DB) *GormProductRepository {
	return &GormProductRepository{db: db}
}

// WithTransaction executes operations within a database transaction

func (r *GormProductRepository) CreateProduct(ctx context.Context, product *entities.Product) error {
	rec := recordFromProduct(product)
	result := r.db.WithContext(ctx).Create(&rec)
	return result.Error
}

func (r *GormProductRepository) GetProductByID(ctx context.Context, id string) (*entities.Product, error) {
	var rec migration.ProductRecord
	result := r.db.WithContext(ctx).First(&rec, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}
	return productFromRecord(rec)
}

func (r *GormProductRepository) UpdateProduct(ctx context.Context, product *entities.Product) error {
	rec := recordFromProduct(product)
	result := r.db.WithContext(ctx).Save(&rec)
	return result.Error
}

func (r *GormProductRepository) DeleteProduct(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&migration.ProductRecord{}, "id = ?", id)
	return result.Error
}

func (r *GormProductRepository) ListProducts(ctx context.Context, page, limit int, search string) ([]*entities.Product, int32, error) {
	var recs []migration.ProductRecord
	var total int64

	query := r.db.WithContext(ctx).Model(&migration.ProductRecord{})

	if search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Find(&recs).Error; err != nil {
		return nil, 0, err
	}

	products := make([]*entities.Product, 0, len(recs))
	for _, rec := range recs {
		product, err := productFromRecord(rec)
		if err != nil {
			return nil, 0, err
		}
		products = append(products, product)
	}

	return products, int32(total), nil
}

func (r *GormProductRepository) ProductExistsByID(ctx context.Context, id string) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&migration.ProductRecord{}).Where("id = ?", id).Count(&count)
	return count > 0, result.Error
}
