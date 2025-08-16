package app

import (
	"context"
	"net"
	"time"

	app "payment-service/internal/app/services"
	srv "payment-service/internal/grpc"
	mockproc "payment-service/internal/infra/processor"
	pb "proto-go/payment"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func Run(ctx context.Context, cfg *Config, logger *zap.Logger) error {
	log := logger.Sugar()

	server := grpc.NewServer()
	processor := mockproc.NewMockPaymentProcessor()
	paymentService := app.NewPaymentService(processor)
	pb.RegisterPaymentServiceServer(server, srv.NewPaymentServer(paymentService))

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Errorw("listen failed", "error", err)
		return err
	}
	log.Infow("payment-service starting", "port", cfg.Port)

	serveErr := make(chan error, 1)
	go func() { serveErr <- server.Serve(lis) }()

	select {
	case <-ctx.Done():
		log.Infow("shutting down payment-service...")
		server.GracefulStop()
		time.Sleep(100 * time.Millisecond)
		return nil
	case err := <-serveErr:
		log.Errorw("serve failed", "error", err)
		return err
	}
}
