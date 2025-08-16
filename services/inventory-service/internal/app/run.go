package app

import (
	"context"
	"net"
	"time"

	appsvc "inventory-service/internal/app/services"
	grpcsvr "inventory-service/internal/infra/grpc"
	repo "inventory-service/internal/infra/repository"
	invpb "proto-go/inventory"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func Run(ctx context.Context, cfg *Config, logger *zap.Logger) error {
	log := logger.Sugar()

	repository := repo.NewInMemoryInventoryRepository()
	svc := appsvc.NewInventoryService(repository)

	grpcServer := grpc.NewServer()
	invpb.RegisterInventoryServiceServer(grpcServer, grpcsvr.NewPBInventoryServer(svc))

	h := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, h)

	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Errorw("listen failed", "error", err)
		return err
	}
	log.Infow("inventory-service starting", "port", cfg.Port)

	serveErr := make(chan error, 1)
	go func() { serveErr <- grpcServer.Serve(lis) }()

	select {
	case <-ctx.Done():
		log.Infow("shutting down inventory-service...")
		grpcServer.GracefulStop()
		time.Sleep(100 * time.Millisecond)
		return nil
	case err := <-serveErr:
		log.Errorw("serve failed", "error", err)
		return err
	}
}
