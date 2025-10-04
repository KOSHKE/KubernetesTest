package repository

import (
	"context"

	"ecommerce-platform/pkg/outbox"
	"ecommerce-platform/services/order-service/internal/domain/aggregates"
	"ecommerce-platform/services/order-service/internal/domain/ports/repository"

	"gorm.io/gorm"
)

// OrderRepositoryFacadeGorm implements the OrderRepositoryFacade interface using GORM
// This facade combines OrderRepository and OutboxRepository functionality
type OrderRepositoryFacadeGorm struct {
	db                *gorm.DB
	orderRepoFactory  func(*gorm.DB) repository.OrderRepository
	outboxRepoFactory func(*gorm.DB) repository.OutboxRepository
}

// NewOrderRepositoryFacade creates a new order repository facade
func NewOrderRepositoryFacade(db *gorm.DB) repository.OrderRepositoryFacade {
	return &OrderRepositoryFacadeGorm{
		db: db,
		orderRepoFactory: func(db *gorm.DB) repository.OrderRepository {
			return NewOrderRepository(db)
		},
		outboxRepoFactory: func(db *gorm.DB) repository.OutboxRepository {
			// Use service-specific outbox table
			return outbox.NewGormRepositoryWithTable(db, "order_outbox_events")
		},
	}
}

// Order operations - delegate to OrderRepository
func (r *OrderRepositoryFacadeGorm) Create(ctx context.Context, order *aggregates.Order) error {
	return r.orderRepoFactory(r.db).Create(ctx, order)
}

func (r *OrderRepositoryFacadeGorm) GetByID(ctx context.Context, id string) (*aggregates.Order, error) {
	return r.orderRepoFactory(r.db).GetByID(ctx, id)
}

func (r *OrderRepositoryFacadeGorm) GetByUserID(ctx context.Context, userID string, page, limit int) ([]*aggregates.Order, int64, error) {
	return r.orderRepoFactory(r.db).GetByUserID(ctx, userID, page, limit)
}

func (r *OrderRepositoryFacadeGorm) Update(ctx context.Context, order *aggregates.Order) error {
	return r.orderRepoFactory(r.db).Update(ctx, order)
}

// Outbox operations - delegate to OutboxRepository
func (r *OrderRepositoryFacadeGorm) SaveEvent(ctx context.Context, event outbox.Event) error {
	return r.outboxRepoFactory(r.db).SaveEvent(ctx, event)
}

func (r *OrderRepositoryFacadeGorm) GetUnprocessedEvents(ctx context.Context, limit int) ([]outbox.Event, error) {
	return r.outboxRepoFactory(r.db).GetUnprocessedEvents(ctx, limit)
}

func (r *OrderRepositoryFacadeGorm) MarkAsProcessed(ctx context.Context, id uint) error {
	return r.outboxRepoFactory(r.db).MarkAsProcessed(ctx, id)
}

func (r *OrderRepositoryFacadeGorm) MarkAsFailed(ctx context.Context, id uint, err string) error {
	return r.outboxRepoFactory(r.db).MarkAsFailed(ctx, id, err)
}

// Transaction support - all repositories participate in the same transaction
func (r *OrderRepositoryFacadeGorm) WithTransaction(ctx context.Context, fn func(repo repository.OrderRepositoryFacade) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &OrderRepositoryFacadeGorm{
			db:                tx,
			orderRepoFactory:  r.orderRepoFactory,
			outboxRepoFactory: r.outboxRepoFactory,
		}
		return fn(txRepo)
	})
}
