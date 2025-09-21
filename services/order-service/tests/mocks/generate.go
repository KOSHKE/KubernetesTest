//go:build tools

package mocks

// Only generate mocks that are actually used in tests
//go:generate go run go.uber.org/mock/mockgen@latest -source=../../../../pkg/logger/logger.go -destination=mock_logger.go -package=mocks
//go:generate go run go.uber.org/mock/mockgen@latest -source=../../../../pkg/outbox/outbox.go -destination=mock_outbox.go -package=mocks
//go:generate go run go.uber.org/mock/mockgen@latest -source=../../internal/domain/ports/repository/order_repository_facade.go -destination=mock_repository.go -package=mocks
