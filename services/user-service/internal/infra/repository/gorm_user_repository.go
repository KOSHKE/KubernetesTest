package repository

import (
	"context"
	"errors"
	"time"

	"user-service/internal/domain/entities"
	"user-service/internal/domain/valueobjects"

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
		FirstName:    u.FirstName(),
		LastName:     u.LastName(),
		Phone:        u.Phone(),
		CreatedAt:    u.CreatedAt(),
		UpdatedAt:    u.UpdatedAt(),
	}
}

func entityFromRecord(r UserRecord) (*entities.User, error) {
	email, err := valueobjects.NewEmail(r.Email)
	if err != nil {
		return nil, err
	}
	password := valueobjects.NewPasswordFromHash(r.PasswordHash)
	return entities.NewUser(r.ID, email, password, r.FirstName, r.LastName, r.Phone), nil
}

type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) Create(ctx context.Context, user *entities.User) error {
	rec := recordFromEntity(user)
	result := r.db.WithContext(ctx).Create(&rec)
	return result.Error
}

func (r *GormUserRepository) GetByID(ctx context.Context, id string) (*entities.User, error) {
	var rec UserRecord
	result := r.db.WithContext(ctx).First(&rec, "id = ?", id)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("user not found")
	}
	return entityFromRecord(rec)
}

func (r *GormUserRepository) GetByEmail(ctx context.Context, email valueobjects.Email) (*entities.User, error) {
	var rec UserRecord
	result := r.db.WithContext(ctx).Where("email = ?", email.Value()).First(&rec)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, errors.New("user not found")
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

// AutoMigrate creates tables
func (r *GormUserRepository) AutoMigrate() error {
	return r.db.AutoMigrate(&UserRecord{})
}
