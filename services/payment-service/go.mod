module payment-service

go 1.25

require (
	github.com/confluentinc/confluent-kafka-go v1.9.2
	github.com/redis/go-redis/v9 v9.6.1
	go.uber.org/zap v1.27.0
	google.golang.org/grpc v1.66.2
	google.golang.org/protobuf v1.34.2
	kubernetetest/pkg/kafka v0.0.0
	proto-go v0.0.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/net v0.26.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240604185151-ef581f913117 // indirect
)

replace proto-go => ../../proto-go

replace kubernetetest/pkg/kafka => ../../pkg/kafka
