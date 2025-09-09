package interfaces

import (
	"ecommerce-platform/services/inventory-service/internal/domain/ports/repository"
)

//go:generate go run go.uber.org/mock/mockgen@latest -source=repository.go -destination=../mocks/mock_repository.go -package=mocks

// InventoryRepositoryFacade interface for testing - same as production interface
type InventoryRepositoryFacade interface {
	repository.InventoryRepositoryFacade
}

// Logger interface for logger in tests
type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
}
