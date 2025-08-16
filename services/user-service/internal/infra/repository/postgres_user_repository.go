package repository

import (
	"context"
	"database/sql"
	"fmt"

	"user-service/internal/domain/entities"
	"user-service/internal/domain/valueobjects"

	_ "github.com/lib/pq"
)

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Save(ctx context.Context, user *entities.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, first_name, last_name, phone, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		user.ID(),
		user.Email().Value(),
		user.Password().HashedValue(),
		user.FirstName(),
		user.LastName(),
		user.Phone(),
		user.CreatedAt(),
		user.UpdatedAt(),
	)

	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	return nil
}

func (r *PostgresUserRepository) FindByID(ctx context.Context, id string) (*entities.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, phone, created_at, updated_at
		FROM users WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)

	var emailStr, passwordHash, firstName, lastName, phone string
	var createdAt, updatedAt string

	err := row.Scan(&id, &emailStr, &passwordHash, &firstName, &lastName, &phone, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to scan user: %w", err)
	}

	email, err := valueobjects.NewEmail(emailStr)
	if err != nil {
		return nil, fmt.Errorf("invalid email in database: %w", err)
	}

	password := valueobjects.NewPasswordFromHash(passwordHash)

	user := entities.NewUser(id, email, password, firstName, lastName, phone)

	return user, nil
}

func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email valueobjects.Email) (*entities.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, phone, created_at, updated_at
		FROM users WHERE email = $1
	`

	row := r.db.QueryRowContext(ctx, query, email.Value())

	var id, emailStr, passwordHash, firstName, lastName, phone string
	var createdAt, updatedAt string

	err := row.Scan(&id, &emailStr, &passwordHash, &firstName, &lastName, &phone, &createdAt, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to scan user: %w", err)
	}

	password := valueobjects.NewPasswordFromHash(passwordHash)

	user := entities.NewUser(id, email, password, firstName, lastName, phone)

	return user, nil
}

func (r *PostgresUserRepository) Update(ctx context.Context, user *entities.User) error {
	query := `
		UPDATE users 
		SET first_name = $2, last_name = $3, phone = $4, updated_at = $5
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query,
		user.ID(),
		user.FirstName(),
		user.LastName(),
		user.Phone(),
		user.UpdatedAt(),
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (r *PostgresUserRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (r *PostgresUserRepository) ExistsByEmail(ctx context.Context, email valueobjects.Email) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, email.Value()).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check user existence: %w", err)
	}

	return exists, nil
}

// InitSchema creates the users table if it doesn't exist
func (r *PostgresUserRepository) InitSchema(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(255) PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			first_name VARCHAR(255),
			last_name VARCHAR(255),
			phone VARCHAR(50),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	`

	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	return nil
}
