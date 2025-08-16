module inventory-service

go 1.25

require (
	go.uber.org/zap v1.27.0
	google.golang.org/grpc v1.66.2
	proto-go v0.0.0
)

require (
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/net v0.26.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240604185151-ef581f913117 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
)

replace proto-go => ../../proto-go
