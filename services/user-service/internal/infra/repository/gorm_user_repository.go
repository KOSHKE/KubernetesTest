package repository

import (
	"context"
	"time"

	"ecommerce-platform/services/user-service/internal/domain/entities"
	"ecommerce-platform/services/user-service/internal/domain/ports/repository"
	"ecommerce-platform/services/user-service/internal/domain/valueobjects"

	"gorm.io/gorm"
)

// UserRecord is a GORM model separated from the domain entity
type UserRecord struct {
	ID           string    `gorm:"primaryKey;type:varchar(255)"`
	Email        string    `gorm:"unique;not null;type:varchar(255)"`
	PasswordHash string    `gorm:"column:password_hash;not null;type:varchar(255)"`
	FirstName    string    `gorm:"not null;type:varchar(255)"`
	LastName     string    `gorm:"not null;type:varchar(255)"`
	Phone        string    `gorm:"type:varchar(50)"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

func (UserRecord) TableName() string { return "users" }

func recordFromEntity(u *entities.User) UserRecord {
	return UserRecord{
		ID:           u.ID(),
		Email:        u.Email().Value(),
		PasswordHash: u.Password().HashedValue(),
		FirstName:    u.FirstName().Value(),
		LastName:     u.LastName().Value(),
		Phone:        u.Phone().Value(),
		CreatedAt:    u.CreatedAt(),
		UpdatedAt:    u.UpdatedAt(),
	}
}

func entityFromRecord(r UserRecord) (*entities.User, error) {
	email := valueobjects.NewEmail(r.Email)
	password := valueobjects.NewPassword(r.PasswordHash)
	firstName := valueobjects.NewName(r.FirstName)
	lastName := valueobjects.NewName(r.LastName)
	phone := valueobjects.NewPhone(r.Phone)
	return entities.NewUser(r.ID, email, password, firstName, lastName, phone), nil
}

type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

// WithTx returns a new repository instance with transaction
func (r *GormUserRepository) WithTx(tx interface{}) repository.UserRepository {
	return &GormUserRepository{db: tx.(*gorm.DB)}
}

func (r *GormUserRepository) Create(ctx context.Context, user *entities.User) error {
	rec := recordFromEntity(user)
	result := r.db.WithContext(ctx).Create(&rec)
	return result.Error
}

func (r *GormUserRepository) GetByID(ctx context.Context, id string) (*entities.User, error) {
	var rec UserRecord
	result := r.db.WithContext(ctx).First(&rec, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}
	return entityFromRecord(rec)
}

func (r *GormUserRepository) GetByEmail(ctx context.Context, email valueobjects.Email) (*entities.User, error) {
	var rec UserRecord
	result := r.db.WithContext(ctx).Where("email = ?", email.Value()).First(&rec)
	if result.Error != nil {
		return nil, result.Error
	}
	return entityFromRecord(rec)
}

func (r *GormUserRepository) Update(ctx context.Context, user *entities.User) error {
	rec := recordFromEntity(user)
	result := r.db.WithContext(ctx).Save(&rec)
	return result.Error
}

func (r *GormUserRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&UserRecord{}, "id = ?", id)
	return result.Error
}

func (r *GormUserRepository) ExistsByEmail(ctx context.Context, email valueobjects.Email) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&UserRecord{}).Where("email = ?", email.Value()).Count(&count)
	return count > 0, result.Error
}

// ExistsByID checks if user exists by ID
func (r *GormUserRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&UserRecord{}).Where("id = ?", id).Count(&count)
	return count > 0, result.Error
}
