//go:build tools

package mocks

//go:generate go run go.uber.org/mock/mockgen@latest -source=../../../../pkg/logger/logger.go -destination=mock_logger.go -package=mocks
//go:generate go run go.uber.org/mock/mockgen@latest -source=../../../../pkg/outbox/outbox.go -destination=mock_outbox.go -package=mocks
//go:generate go run go.uber.org/mock/mockgen@latest -source=../../../../pkg/kafkaclient/publisher.go -destination=mock_kafka_publisher.go -package=mocks
//go:generate go run go.uber.org/mock/mockgen@latest -source=../../internal/domain/ports/repository/inventory_repository_facade.go -destination=mock_repository.go -package=mocks
//go:generate go run go.uber.org/mock/mockgen@latest -source=../../internal/domain/ports/publisher/event_publisher.go -destination=mock_publisher.go -package=mocks
