//go:build tools

package mocks

//go:generate go run go.uber.org/mock/mockgen@latest -source=../../../../pkg/logger/logger.go -destination=mock_logger.go -package=mocks
//go:generate go run go.uber.org/mock/mockgen@latest -source=../../../../pkg/outbox/outbox.go -destination=mock_outbox.go -package=mocks
//go:generate go run go.uber.org/mock/mockgen@latest -source=../../../../pkg/kafkaclient/publisher.go -destination=mock_kafka_publisher.go -package=mocks
//go:generate go run go.uber.org/mock/mockgen@latest -source=../../internal/domain/ports/repository/outbox_repository.go -destination=mock_outbox_repository.go -package=mocks
//go:generate go run go.uber.org/mock/mockgen@latest -source=../../internal/domain/ports/publisher/payment_events_publisher.go -destination=mock_payment_events_publisher.go -package=mocks
//go:generate go run go.uber.org/mock/mockgen@latest -source=../../internal/metrics/payment_metrics.go -destination=mock_payment_metrics.go -package=mocks
