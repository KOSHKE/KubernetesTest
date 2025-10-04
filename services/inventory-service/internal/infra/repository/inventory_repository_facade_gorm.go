package repository

import (
	"context"

	"ecommerce-platform/pkg/outbox"
	"ecommerce-platform/services/inventory-service/internal/domain/aggregates"
	"ecommerce-platform/services/inventory-service/internal/domain/entities"
	"ecommerce-platform/services/inventory-service/internal/domain/ports/repository"

	"gorm.io/gorm"
)

// InventoryRepositoryFacadeGorm implements the InventoryRepository interface using GORM
// This facade combines ProductRepository, StockRepository and OutboxRepository functionality
type InventoryRepositoryFacadeGorm struct {
	db                 *gorm.DB
	productRepoFactory func(*gorm.DB) repository.ProductRepository
	stockRepoFactory   func(*gorm.DB) repository.StockRepository
	outboxRepoFactory  func(*gorm.DB) repository.OutboxRepository
}

// NewInventoryRepository creates a new inventory repository
func NewInventoryRepository(db *gorm.DB) repository.InventoryRepositoryFacade {
	return &InventoryRepositoryFacadeGorm{
		db: db,
		productRepoFactory: func(db *gorm.DB) repository.ProductRepository {
			return &GormProductRepository{db: db}
		},
		stockRepoFactory: func(db *gorm.DB) repository.StockRepository {
			return &GormStockRepository{db: db}
		},
		outboxRepoFactory: func(db *gorm.DB) repository.OutboxRepository {
			// Use service-specific outbox table for inventory service
			return outbox.NewGormRepositoryWithTable(db, "inventory_outbox_events")
		},
	}
}

// Product operations - automatically inherited from ProductRepository via interface composition
func (r *InventoryRepositoryFacadeGorm) CreateProduct(ctx context.Context, product *entities.Product) error {
	return r.productRepoFactory(r.db).CreateProduct(ctx, product)
}

func (r *InventoryRepositoryFacadeGorm) GetProductByID(ctx context.Context, id string) (*entities.Product, error) {
	return r.productRepoFactory(r.db).GetProductByID(ctx, id)
}

func (r *InventoryRepositoryFacadeGorm) UpdateProduct(ctx context.Context, product *entities.Product) error {
	return r.productRepoFactory(r.db).UpdateProduct(ctx, product)
}

func (r *InventoryRepositoryFacadeGorm) DeleteProduct(ctx context.Context, id string) error {
	return r.productRepoFactory(r.db).DeleteProduct(ctx, id)
}

func (r *InventoryRepositoryFacadeGorm) ListProducts(ctx context.Context, page, limit int, search string) ([]*entities.Product, int32, error) {
	return r.productRepoFactory(r.db).ListProducts(ctx, page, limit, search)
}

func (r *InventoryRepositoryFacadeGorm) ProductExistsByID(ctx context.Context, id string) (bool, error) {
	return r.productRepoFactory(r.db).ProductExistsByID(ctx, id)
}

// ListProductsWithStock retrieves products with their stock information using a single query
func (r *InventoryRepositoryFacadeGorm) ListProductsWithStock(ctx context.Context, page, limit int, search string) ([]*aggregates.ProductInventory, int32, error) {
	return r.productRepoFactory(r.db).ListProductsWithStock(ctx, page, limit, search)
}

// Stock operations - automatically inherited from StockRepository via interface composition

func (r *InventoryRepositoryFacadeGorm) GetStockByID(ctx context.Context, id string) (*entities.Stock, error) {
	return r.stockRepoFactory(r.db).GetStockByID(ctx, id)
}

func (r *InventoryRepositoryFacadeGorm) StockExistsByID(ctx context.Context, id string) (bool, error) {
	return r.stockRepoFactory(r.db).StockExistsByID(ctx, id)
}

func (r *InventoryRepositoryFacadeGorm) GetStockByProductID(ctx context.Context, productID string) (*entities.Stock, error) {
	return r.stockRepoFactory(r.db).GetStockByProductID(ctx, productID)
}

func (r *InventoryRepositoryFacadeGorm) UpsertStock(ctx context.Context, stock *entities.Stock) error {
	return r.stockRepoFactory(r.db).UpsertStock(ctx, stock)
}

func (r *InventoryRepositoryFacadeGorm) GetStocksByProductIDs(ctx context.Context, productIDs []string, forUpdate bool) (map[string]*entities.Stock, error) {
	return r.stockRepoFactory(r.db).GetStocksByProductIDs(ctx, productIDs, forUpdate)
}

func (r *InventoryRepositoryFacadeGorm) UpsertStocks(ctx context.Context, stocks []*entities.Stock) error {
	return r.stockRepoFactory(r.db).UpsertStocks(ctx, stocks)
}

// Outbox operations - automatically inherited from OutboxRepository via interface composition
func (r *InventoryRepositoryFacadeGorm) SaveEvent(ctx context.Context, event outbox.Event) error {
	return r.outboxRepoFactory(r.db).SaveEvent(ctx, event)
}

func (r *InventoryRepositoryFacadeGorm) GetUnprocessedEvents(ctx context.Context, limit int) ([]outbox.Event, error) {
	return r.outboxRepoFactory(r.db).GetUnprocessedEvents(ctx, limit)
}

func (r *InventoryRepositoryFacadeGorm) MarkAsProcessed(ctx context.Context, id uint) error {
	return r.outboxRepoFactory(r.db).MarkAsProcessed(ctx, id)
}

func (r *InventoryRepositoryFacadeGorm) MarkAsFailed(ctx context.Context, id uint, err string) error {
	return r.outboxRepoFactory(r.db).MarkAsFailed(ctx, id, err)
}

// Transaction support - all repositories participate in the same transaction
func (r *InventoryRepositoryFacadeGorm) WithTransaction(ctx context.Context, fn func(repo repository.InventoryRepositoryFacade) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &InventoryRepositoryFacadeGorm{
			db:                 tx,
			productRepoFactory: r.productRepoFactory,
			stockRepoFactory:   r.stockRepoFactory,
			outboxRepoFactory:  r.outboxRepoFactory,
		}
		return fn(txRepo)
	})
}
