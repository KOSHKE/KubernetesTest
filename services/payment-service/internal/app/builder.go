package app

import (
	"ecommerce-platform/pkg/config"
	"ecommerce-platform/pkg/logger"
	"ecommerce-platform/pkg/metrics"
	"ecommerce-platform/services/payment-service/internal/domain/ports/grpc"
	"ecommerce-platform/services/payment-service/internal/domain/ports/publisher"
	paymentGrpc "ecommerce-platform/services/payment-service/internal/infra/grpc"
	publisherimpl "ecommerce-platform/services/payment-service/internal/infra/publisher"
	paymentmetrics "ecommerce-platform/services/payment-service/internal/metrics"

	grpcLib "google.golang.org/grpc"
)

// Builder constructs application dependencies
type Builder struct {
	config *config.PaymentConfig
	logger logger.Logger
}

// NewBuilder creates a new application builder
func NewBuilder(cfg *config.PaymentConfig, logger logger.Logger) *Builder {
	return &Builder{
		config: cfg,
		logger: logger,
	}
}

// BuildPaymentProcessedPublisher creates payment processed publisher
func (b *Builder) BuildPaymentProcessedPublisher() (publisher.PaymentProcessedPublisher, error) {
	return publisherimpl.NewPaymentProcessedPublisher(
		b.config.GetKafkaBrokers(),
		"payments.v1.payment_processed",
	)
}

// BuildPaymentGrpcServer creates payment gRPC server
func (b *Builder) BuildPaymentGrpcServer() grpc.PaymentServer {
	return paymentGrpc.NewPaymentServerImpl()
}

// BuildMetrics creates payment metrics
func (b *Builder) BuildMetrics() paymentmetrics.PaymentMetrics {
	return paymentmetrics.NewPaymentMetrics()
}

// BuildMetricsServer creates metrics server
func (b *Builder) BuildMetricsServer() *metrics.MetricsServer {
	return metrics.NewMetricsServer(":"+b.config.MetricsPort, b.logger)
}

// BuildGrpcServer creates gRPC server
func (b *Builder) BuildGrpcServer() *grpcLib.Server {
	return grpcLib.NewServer()
}
