//go:build tools

package mocks

//go:generate go run go.uber.org/mock/mockgen@latest -source=../../../../pkg/logger/logger.go -destination=mock_logger.go -package=mocks
//go:generate go run go.uber.org/mock/mockgen@latest -source=../../internal/domain/ports/repository/user_repository.go -destination=mock_user_repository.go -package=mocks
//go:generate go run go.uber.org/mock/mockgen@latest -source=../../internal/domain/ports/repository/session_repository.go -destination=mock_session_repository.go -package=mocks
//go:generate go run go.uber.org/mock/mockgen@latest -source=../../internal/domain/ports/services/token_generator.go -destination=mock_token_generator.go -package=mocks
