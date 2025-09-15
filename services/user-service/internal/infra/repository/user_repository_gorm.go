package repository

import (
	"context"
	"time"

	"ecommerce-platform/services/user-service/internal/domain/entities"
	"ecommerce-platform/services/user-service/internal/domain/ports/repository"
	"ecommerce-platform/services/user-service/internal/domain/valueobjects"

	"gorm.io/gorm"
)

// userRecord is a GORM model separated from the domain entity
type userRecord struct {
	ID           string    `gorm:"primaryKey;type:varchar(255)"`
	Email        string    `gorm:"unique;not null;type:varchar(255)"`
	PasswordHash string    `gorm:"column:password_hash;not null;type:varchar(255)"`
	FirstName    string    `gorm:"not null;type:varchar(255)"`
	LastName     string    `gorm:"not null;type:varchar(255)"`
	Phone        string    `gorm:"type:varchar(50)"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

func (userRecord) TableName() string { return "users" }

func recordFromEntity(u *entities.User) userRecord {
	return userRecord{
		ID:           u.ID,
		Email:        u.Email.Value(),
		PasswordHash: u.Password.HashedValue(),
		FirstName:    u.FirstName.Value(),
		LastName:     u.LastName.Value(),
		Phone:        u.Phone.Value(),
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

func entityFromRecord(r userRecord) (*entities.User, error) {
	email, err := valueobjects.NewEmail(r.Email)
	if err != nil {
		return nil, err
	}

	password := valueobjects.NewPasswordFromHashed(r.PasswordHash)
	firstName := valueobjects.NewName(r.FirstName)
	lastName := valueobjects.NewName(r.LastName)

	phone, err := valueobjects.NewPhone(r.Phone)
	if err != nil {
		return nil, err
	}

	return entities.NewUser(r.ID, email, password, firstName, lastName, phone, r.CreatedAt, r.UpdatedAt), nil
}

type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

// WithTransaction executes operations within a database transaction
func (r *GormUserRepository) WithTransaction(ctx context.Context, fn func(repository.UserRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &GormUserRepository{db: tx}
		return fn(txRepo)
	})
}

func (r *GormUserRepository) Create(ctx context.Context, user *entities.User) error {
	rec := recordFromEntity(user)
	result := r.db.WithContext(ctx).Create(&rec)
	return result.Error
}

func (r *GormUserRepository) GetByID(ctx context.Context, id string) (*entities.User, error) {
	var rec userRecord
	result := r.db.WithContext(ctx).First(&rec, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}
	return entityFromRecord(rec)
}

func (r *GormUserRepository) GetByEmail(ctx context.Context, email valueobjects.Email) (*entities.User, error) {
	var rec userRecord
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
	result := r.db.WithContext(ctx).Delete(&userRecord{}, "id = ?", id)
	return result.Error
}

func (r *GormUserRepository) ExistsByEmail(ctx context.Context, email valueobjects.Email) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&userRecord{}).Where("email = ?", email.Value()).Count(&count)
	return count > 0, result.Error
}

// ExistsByID checks if user exists by ID
func (r *GormUserRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&userRecord{}).Where("id = ?", id).Count(&count)
	return count > 0, result.Error
}
