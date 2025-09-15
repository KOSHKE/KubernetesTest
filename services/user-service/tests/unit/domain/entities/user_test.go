package entities_test

import (
	"testing"
	"time"

	"ecommerce-platform/services/user-service/internal/domain/entities"
	"ecommerce-platform/services/user-service/internal/domain/valueobjects"

	"github.com/stretchr/testify/assert"
)

func TestUser_VerifyPassword(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		// Arrange
		user := createTestUser()
		plainPassword := "TestPassword123!"

		// Act
		result := user.VerifyPassword(plainPassword)

		// Assert
		assert.True(t, result)
	})

	t.Run("wrong password", func(t *testing.T) {
		t.Parallel()
		// Arrange
		user := createTestUser()
		wrongPassword := "WrongPassword123!"

		// Act
		result := user.VerifyPassword(wrongPassword)

		// Assert
		assert.False(t, result)
	})

	t.Run("empty password", func(t *testing.T) {
		t.Parallel()
		// Arrange
		user := createTestUser()
		emptyPassword := ""

		// Act
		result := user.VerifyPassword(emptyPassword)

		// Assert
		assert.False(t, result)
	})
}

func TestUser_NewUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		// Arrange
		id := "user-123"
		email, _ := valueobjects.NewEmail("test@example.com")
		password, _ := valueobjects.NewPassword("TestPassword123!")
		firstName := valueobjects.NewName("John")
		lastName := valueobjects.NewName("Doe")
		phone, _ := valueobjects.NewPhone("+1234567890")
		createdAt := time.Now()
		updatedAt := time.Now()

		// Act
		user := entities.NewUser(id, email, password, firstName, lastName, phone, createdAt, updatedAt)

		// Assert
		assert.Equal(t, id, user.ID)
		assert.Equal(t, email, user.Email)
		assert.Equal(t, password, user.Password)
		assert.Equal(t, firstName, user.FirstName)
		assert.Equal(t, lastName, user.LastName)
		assert.Equal(t, phone, user.Phone)
		assert.Equal(t, createdAt, user.CreatedAt)
		assert.Equal(t, updatedAt, user.UpdatedAt)
	})
}

// Helper functions for creating test data
func createTestUser() *entities.User {
	id := "user-123"
	email, _ := valueobjects.NewEmail("test@example.com")
	password, _ := valueobjects.NewPassword("TestPassword123!")
	firstName := valueobjects.NewName("John")
	lastName := valueobjects.NewName("Doe")
	phone, _ := valueobjects.NewPhone("+1234567890")
	createdAt := time.Now()
	updatedAt := time.Now()

	return entities.NewUser(id, email, password, firstName, lastName, phone, createdAt, updatedAt)
}
