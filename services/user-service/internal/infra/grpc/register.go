package grpc

import (
	userpb "ecommerce-platform/proto-go/user"
	"ecommerce-platform/services/user-service/internal/application/services"

	gogrpc "google.golang.org/grpc"
)

// RegisterUserPBServer hides proto dependency from main
func RegisterUserPBServer(server *gogrpc.Server, svc *services.UserApplicationService) {
	pbServer := NewPBUserServer(svc)
	userpb.RegisterUserServiceServer(server, pbServer)
}
