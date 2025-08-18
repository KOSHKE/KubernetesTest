package services

import (
	"context"
	"inventory-service/internal/domain/models"
	"inventory-service/internal/ports/clock"
	invpub "inventory-service/internal/ports/events"
	"inventory-service/internal/ports/idgen"
	"inventory-service/internal/ports/repository"
	"sync"
	"time"

	events "proto-go/events"
)

type InventoryService struct {
	repo  repository.InventoryRepository
	clock clock.Clock
	ids   idgen.IDGenerator
	pub   invpub.Publisher

	mu              sync.Mutex
	reservedByOrder map[string][]StockCheckItem
}

func NewInventoryService(repo repository.InventoryRepository) *InventoryService {
	return &InventoryService{repo: repo, reservedByOrder: make(map[string][]StockCheckItem)}
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

// GetStockQuantity returns available quantity for product
func (s *InventoryService) GetStockQuantity(ctx context.Context, productID string) (int32, error) {
	st, err := s.repo.GetStock(ctx, productID)
	if err != nil {
		return 0, err
	}
	return st.AvailableQuantity, nil
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
	// store reservation for later commit/release
	s.mu.Lock()
	s.reservedByOrder[orderID] = items
	s.mu.Unlock()

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

// FinalizeReservation commits (success=true) or releases (success=false) reserved stock for order
func (s *InventoryService) FinalizeReservation(ctx context.Context, orderID string, success bool) {
	s.mu.Lock()
	items := s.reservedByOrder[orderID]
	delete(s.reservedByOrder, orderID)
	s.mu.Unlock()
	if len(items) == 0 {
		return
	}
	for _, it := range items {
		if success {
			_ = s.repo.Commit(ctx, it.ProductID, it.Quantity)
		} else {
			_ = s.repo.Release(ctx, it.ProductID, it.Quantity)
		}
	}
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
