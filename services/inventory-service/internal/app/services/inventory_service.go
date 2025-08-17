package services

import (
	"context"
	"inventory-service/internal/domain/models"
	"inventory-service/internal/ports/clock"
	invpub "inventory-service/internal/ports/events"
	"inventory-service/internal/ports/idgen"
	"inventory-service/internal/ports/repository"
	"time"

	events "proto-go/events"
)

type InventoryService struct {
	repo  repository.InventoryRepository
	clock clock.Clock
	ids   idgen.IDGenerator
	pub   invpub.Publisher
}

func NewInventoryService(repo repository.InventoryRepository) *InventoryService {
	return &InventoryService{repo: repo}
}

// WithClock allows injecting custom clock
func (s *InventoryService) WithClock(c clock.Clock) *InventoryService { s.clock = c; return s }

// WithIDGenerator allows injecting custom id generator
func (s *InventoryService) WithIDGenerator(g idgen.IDGenerator) *InventoryService {
	s.ids = g
	return s
}

// WithPublisher sets event publisher
func (s *InventoryService) WithPublisher(p invpub.Publisher) *InventoryService { s.pub = p; return s }

type StockCheckItem struct {
	ProductID string
	Quantity  int32
}

type StockCheckResult struct {
	ProductID         string
	RequestedQuantity int32
	AvailableQuantity int32
	IsAvailable       bool
}

func (s *InventoryService) GetProduct(ctx context.Context, id string) (*models.Product, error) {
	return s.repo.GetProduct(ctx, id)
}

func (s *InventoryService) ListProducts(ctx context.Context, categoryID string, page, limit int, search string) ([]*models.Product, int32, error) {
	return s.repo.ListProducts(ctx, categoryID, page, limit, search)
}

func (s *InventoryService) CheckStock(ctx context.Context, items []StockCheckItem) ([]StockCheckResult, bool, error) {
	results := make([]StockCheckResult, 0, len(items))
	allAvailable := true
	for _, it := range items {
		st, err := s.repo.GetStock(ctx, it.ProductID)
		if err != nil {
			results = append(results, StockCheckResult{ProductID: it.ProductID, RequestedQuantity: it.Quantity, AvailableQuantity: 0, IsAvailable: false})
			allAvailable = false
			continue
		}
		ok := st.AvailableQuantity >= it.Quantity
		if !ok {
			allAvailable = false
		}
		results = append(results, StockCheckResult{ProductID: it.ProductID, RequestedQuantity: it.Quantity, AvailableQuantity: st.AvailableQuantity, IsAvailable: ok})
	}
	return results, allAvailable, nil
}

func (s *InventoryService) ReserveStock(ctx context.Context, orderID string, items []StockCheckItem) (failed []string, err error) {
	var failedProducts []string
	for _, it := range items {
		if err := s.repo.Reserve(ctx, it.ProductID, it.Quantity); err != nil {
			failedProducts = append(failedProducts, it.ProductID)
		}
	}
	// could write reservation audit with s.clock.Now()/s.ids.NewID(...)
	_ = s.now()
	_ = s.newID("res-")
	// publish result (best-effort)
	if s.pub != nil {
		if len(failedProducts) == 0 {
			_ = s.pub.PublishStockReserved(ctx, &events.StockReserved{OrderId: orderID, OccurredAt: time.Now().Format(time.RFC3339)})
		} else {
			_ = s.pub.PublishStockReservationFailed(ctx, &events.StockReservationFailed{OrderId: orderID, Reason: "insufficient stock", OccurredAt: time.Now().Format(time.RFC3339)})
		}
	}
	return failedProducts, nil
}

func (s *InventoryService) ReleaseStock(ctx context.Context, orderID string, items []StockCheckItem) error {
	for _, it := range items {
		_ = s.repo.Release(ctx, it.ProductID, it.Quantity)
	}
	_ = s.now()
	return nil
}

func (s *InventoryService) now() int64 {
	if s.clock != nil {
		return s.clock.Now().Unix()
	}
	return 0
}

func (s *InventoryService) newID(prefix string) string {
	if s.ids != nil {
		return s.ids.NewID(prefix)
	}
	return ""
}
