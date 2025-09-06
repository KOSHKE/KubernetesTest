package repository

import (
	"context"

	"ecommerce-platform/services/inventory-service/internal/domain/entities"
	"ecommerce-platform/services/inventory-service/internal/domain/ports/repository"

	"gorm.io/gorm"
)

// InventoryRepositoryImpl implements the InventoryRepository interface
// This facade combines ProductRepository and StockRepository functionality
type InventoryRepositoryImpl struct {
	db                 *gorm.DB
	productRepoFactory func(*gorm.DB) repository.ProductRepository
	stockRepoFactory   func(*gorm.DB) repository.StockRepository
}

// NewInventoryRepository creates a new inventory repository
func NewInventoryRepository(db *gorm.DB) repository.InventoryRepository {
	return &InventoryRepositoryImpl{
		db: db,
		productRepoFactory: func(db *gorm.DB) repository.ProductRepository {
			return &GormProductRepository{db: db}
		},
		stockRepoFactory: func(db *gorm.DB) repository.StockRepository {
			return &GormStockRepository{db: db}
		},
	}
}

// Product operations - automatically inherited from ProductRepository via interface composition
func (r *InventoryRepositoryImpl) CreateProduct(ctx context.Context, product *entities.Product) error {
	return r.productRepoFactory(r.db).CreateProduct(ctx, product)
}

func (r *InventoryRepositoryImpl) GetProductByID(ctx context.Context, id string) (*entities.Product, error) {
	return r.productRepoFactory(r.db).GetProductByID(ctx, id)
}

func (r *InventoryRepositoryImpl) UpdateProduct(ctx context.Context, product *entities.Product) error {
	return r.productRepoFactory(r.db).UpdateProduct(ctx, product)
}

func (r *InventoryRepositoryImpl) DeleteProduct(ctx context.Context, id string) error {
	return r.productRepoFactory(r.db).DeleteProduct(ctx, id)
}

func (r *InventoryRepositoryImpl) ListProducts(ctx context.Context, page, limit int, search string) ([]*entities.Product, int32, error) {
	return r.productRepoFactory(r.db).ListProducts(ctx, page, limit, search)
}

func (r *InventoryRepositoryImpl) ProductExistsByID(ctx context.Context, id string) (bool, error) {
	return r.productRepoFactory(r.db).ProductExistsByID(ctx, id)
}

// Stock operations - automatically inherited from StockRepository via interface composition
func (r *InventoryRepositoryImpl) CreateStock(ctx context.Context, stock *entities.Stock) error {
	return r.stockRepoFactory(r.db).CreateStock(ctx, stock)
}

func (r *InventoryRepositoryImpl) GetStockByID(ctx context.Context, id string) (*entities.Stock, error) {
	return r.stockRepoFactory(r.db).GetStockByID(ctx, id)
}

func (r *InventoryRepositoryImpl) UpdateStock(ctx context.Context, stock *entities.Stock) error {
	return r.stockRepoFactory(r.db).UpdateStock(ctx, stock)
}

func (r *InventoryRepositoryImpl) StockExistsByID(ctx context.Context, id string) (bool, error) {
	return r.stockRepoFactory(r.db).StockExistsByID(ctx, id)
}

func (r *InventoryRepositoryImpl) GetStockByProductID(ctx context.Context, productID string) (*entities.Stock, error) {
	return r.stockRepoFactory(r.db).GetStockByProductID(ctx, productID)
}

// Transaction support - both repositories participate in the same transaction
func (r *InventoryRepositoryImpl) WithTransaction(ctx context.Context, fn func(repo repository.InventoryRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &InventoryRepositoryImpl{
			db:                 tx,
			productRepoFactory: r.productRepoFactory,
			stockRepoFactory:   r.stockRepoFactory,
		}
		return fn(txRepo)
	})
}
