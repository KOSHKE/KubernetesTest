package repository

import (
	"context"
	"user-service/internal/domain/entities"
	"user-service/internal/domain/valueobjects"
)

type UserRepository interface {
	Create(ctx context.Context, user *entities.User) error
	GetByID(ctx context.Context, id string) (*entities.User, error)
	GetByEmail(ctx context.Context, email valueobjects.Email) (*entities.User, error)
	Update(ctx context.Context, user *entities.User) error
	Delete(ctx context.Context, id string) error
	ExistsByEmail(ctx context.Context, email valueobjects.Email) (bool, error)
}
