package grpc

import (
	userpb "proto-go/user"
	"user-service/internal/app/services"

	gogrpc "google.golang.org/grpc"
)

// RegisterUserPBServer hides proto dependency from main
func RegisterUserPBServer(server *gogrpc.Server, svc *services.UserService) {
	pbServer := NewPBUserServer(svc)
	userpb.RegisterUserServiceServer(server, pbServer)
}
