package grpc

import (
	userpb "github.com/kubernetestest/ecommerce-platform/proto-go/user"
	"github.com/kubernetestest/ecommerce-platform/services/user-service/internal/app/services"

	gogrpc "google.golang.org/grpc"
)

// RegisterUserPBServer hides proto dependency from main
func RegisterUserPBServer(server *gogrpc.Server, svc *services.UserService) {
	pbServer := NewPBUserServer(svc)
	userpb.RegisterUserServiceServer(server, pbServer)
}
