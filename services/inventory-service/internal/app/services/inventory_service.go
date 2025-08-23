package services

import (
	"context"
	"errors"
	"github.com/kubernetestest/ecommerce-platform/services/inventory-service/internal/domain/models"
	invpub "github.com/kubernetestest/ecommerce-platform/services/inventory-service/internal/ports/events"
	"github.com/kubernetestest/ecommerce-platform/services/inventory-service/internal/ports/repository"
	"sync"
	"time"

	events "github.com/kubernetestest/ecommerce-platform/proto-go/events"

	"go.uber.org/multierr"
)

type InventoryService struct {
	repo repository.InventoryRepository
	pub  invpub.Publisher

	mu              sync.Mutex
	reservedByOrder map[string][]StockCheckItem
	reservationTTL  time.Duration
	reservedTimers  map[string]*time.Timer
}

func NewInventoryService(repo repository.InventoryRepository) *InventoryService {
	return &InventoryService{repo: repo, reservedByOrder: make(map[string][]StockCheckItem), reservedTimers: make(map[string]*time.Timer)}
}

// WithPublisher sets event publisher
func (s *InventoryService) WithPublisher(p invpub.Publisher) *InventoryService { s.pub = p; return s }

// WithReservationTTL sets TTL for automatic reservation expiration
func (s *InventoryService) WithReservationTTL(d time.Duration) *InventoryService {
	s.reservationTTL = d
	return s
}

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

// GetStocksByIDs returns stocks map for product IDs
func (s *InventoryService) GetStocksByIDs(ctx context.Context, productIDs []string) (map[string]*models.Stock, error) {
	return s.repo.GetStocksByIDs(ctx, productIDs)
}

func (s *InventoryService) GetCategories(ctx context.Context, activeOnly bool) ([]*models.Category, error) {
	return s.repo.GetCategories(ctx, activeOnly)
}

func (s *InventoryService) CheckStock(ctx context.Context, items []StockCheckItem) ([]StockCheckResult, bool, error) {
	results := make([]StockCheckResult, 0, len(items))
	allAvailable := true
	// build unique set of product IDs
	unique := make(map[string]struct{}, len(items))
	order := make([]string, 0, len(items))
	for _, it := range items {
		if _, ok := unique[it.ProductID]; !ok {
			unique[it.ProductID] = struct{}{}
			order = append(order, it.ProductID)
		}
	}
	stocks, err := s.repo.GetStocksByIDs(ctx, order)
	if err != nil {
		return nil, false, err
	}
	for _, it := range items {
		st, ok := stocks[it.ProductID]
		if !ok {
			results = append(results, StockCheckResult{ProductID: it.ProductID, RequestedQuantity: it.Quantity, AvailableQuantity: 0, IsAvailable: false})
			allAvailable = false
			continue
		}
		okAvail := st.AvailableQuantity >= it.Quantity
		if !okAvail {
			allAvailable = false
		}
		results = append(results, StockCheckResult{ProductID: it.ProductID, RequestedQuantity: it.Quantity, AvailableQuantity: st.AvailableQuantity, IsAvailable: okAvail})
	}
	return results, allAvailable, nil
}

func (s *InventoryService) ReserveStock(ctx context.Context, orderID string, userID string, items []StockCheckItem) (failed []string, err error) {
	// idempotency: prevent double reservation for same order
	s.mu.Lock()
	if _, exists := s.reservedByOrder[orderID]; exists {
		s.mu.Unlock()
		return nil, errors.New("order already has an active reservation")
	}
	s.mu.Unlock()

	var failedProducts []string
	successful := make([]StockCheckItem, 0, len(items))
	var aggErr error
	for _, it := range items {
		if rerr := s.repo.Reserve(ctx, it.ProductID, it.Quantity); rerr != nil {
			failedProducts = append(failedProducts, it.ProductID)
			aggErr = multierr.Append(aggErr, rerr)
			continue
		}
		successful = append(successful, it)
	}

	// store only successfully reserved items for later commit/release
	if len(successful) > 0 {
		s.mu.Lock()
		s.reservedByOrder[orderID] = successful
		// schedule TTL expiration if configured
		if s.reservationTTL > 0 {
			if t, ok := s.reservedTimers[orderID]; ok {
				t.Stop()
			}
			s.reservedTimers[orderID] = time.AfterFunc(s.reservationTTL, func() {
				s.expireReservation(context.Background(), orderID)
			})
		}
		s.mu.Unlock()
	}

	// publish result (best-effort)
	if s.pub != nil {
		if len(failedProducts) == 0 {
			_ = s.pub.PublishStockReserved(ctx, &events.StockReserved{OrderId: orderID, UserId: userID, OccurredAt: time.Now().Format(time.RFC3339)})
		} else {
			_ = s.pub.PublishStockReservationFailed(ctx, &events.StockReservationFailed{OrderId: orderID, UserId: userID, Reason: "insufficient stock", OccurredAt: time.Now().Format(time.RFC3339)})
		}
	}
	return failedProducts, aggErr
}

// FinalizeReservation commits (success=true) or releases (success=false) reserved stock for order
func (s *InventoryService) FinalizeReservation(ctx context.Context, orderID string, success bool) error {
	// get and delete reservation under lock
	s.mu.Lock()
	items := s.reservedByOrder[orderID]
	delete(s.reservedByOrder, orderID)
	if t, ok := s.reservedTimers[orderID]; ok {
		t.Stop()
		delete(s.reservedTimers, orderID)
	}
	s.mu.Unlock()
	if len(items) == 0 {
		return nil
	}
	var aggErr error
	for _, it := range items {
		if success {
			if err := s.repo.Commit(ctx, it.ProductID, it.Quantity); err != nil {
				aggErr = multierr.Append(aggErr, err)
			}
		} else {
			if err := s.repo.Release(ctx, it.ProductID, it.Quantity); err != nil {
				aggErr = multierr.Append(aggErr, err)
			}
		}
	}
	return aggErr
}

func (s *InventoryService) ReleaseStock(ctx context.Context, orderID string, items []StockCheckItem) error {
	var aggErr error
	for _, it := range items {
		if err := s.repo.Release(ctx, it.ProductID, it.Quantity); err != nil {
			aggErr = multierr.Append(aggErr, err)
		}
	}
	return aggErr
}

func (s *InventoryService) expireReservation(ctx context.Context, orderID string) {
	// release only if still present (not finalized)
	s.mu.Lock()
	items, ok := s.reservedByOrder[orderID]
	if ok {
		delete(s.reservedByOrder, orderID)
	}
	delete(s.reservedTimers, orderID)
	s.mu.Unlock()
	if !ok || len(items) == 0 {
		return
	}
	for _, it := range items {
		_ = s.repo.Release(ctx, it.ProductID, it.Quantity)
	}
}
