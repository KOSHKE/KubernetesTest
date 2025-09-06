package repository

import (
	"context"

	"ecommerce-platform/services/user-service/internal/domain/entities"
	"ecommerce-platform/services/user-service/internal/domain/valueobjects"
)

type UserRepository interface {
	Create(ctx context.Context, user *entities.User) error
	GetByID(ctx context.Context, id string) (*entities.User, error)
	GetByEmail(ctx context.Context, email valueobjects.Email) (*entities.User, error)
	Update(ctx context.Context, user *entities.User) error
	Delete(ctx context.Context, id string) error
	ExistsByEmail(ctx context.Context, email valueobjects.Email) (bool, error)
	ExistsByID(ctx context.Context, id string) (bool, error)

	// Transaction support
	WithTransaction(ctx context.Context, fn func(UserRepository) error) error
}
