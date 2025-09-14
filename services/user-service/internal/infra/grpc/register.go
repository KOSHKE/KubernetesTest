package grpc

import (
	userpb "ecommerce-platform/proto-go/user"
	"ecommerce-platform/services/user-service/internal/application/services"
	"ecommerce-platform/services/user-service/internal/metrics"

	gogrpc "google.golang.org/grpc"
)

// RegisterUserPBServer hides proto dependency from main
func RegisterUserPBServer(server *gogrpc.Server, svc *services.UserApplicationService, metrics metrics.UserMetrics) {
	pbServer := NewPBUserServer(svc, metrics)
	userpb.RegisterUserServiceServer(server, pbServer)
}
