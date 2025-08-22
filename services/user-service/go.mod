module user-service

go 1.25

require (
	go.uber.org/zap v1.27.0
	golang.org/x/crypto v0.27.0
	google.golang.org/grpc v1.66.2
	google.golang.org/protobuf v1.34.2
	gorm.io/driver/postgres v1.5.9
	gorm.io/gorm v1.25.12
	jwt v0.0.0
	proto-go v0.0.0
	redis v0.0.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/golang-jwt/jwt/v5 v5.2.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.5.5 // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/redis/go-redis/v9 v9.3.1 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1 // indirect
)

replace (
	jwt => ../../libs/jwt
	proto-go => ../../proto-go
	redis => ../../libs/redis
)
