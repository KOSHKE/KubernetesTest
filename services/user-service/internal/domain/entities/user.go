package entities

import (
	"time"
	"user-service/internal/domain/valueobjects"
)

// User represents the User aggregate root
type User struct {
	id        string
	email     valueobjects.Email
	password  valueobjects.Password
	profile   *Profile
	createdAt time.Time
	updatedAt time.Time
}

type Profile struct {
	firstName string
	lastName  string
	phone     string
}

// NewUser creates a new User aggregate
func NewUser(id string, email valueobjects.Email, password valueobjects.Password, firstName, lastName, phone string) *User {
	now := time.Now()
	return &User{
		id:       id,
		email:    email,
		password: password,
		profile: &Profile{
			firstName: firstName,
			lastName:  lastName,
			phone:     phone,
		},
		createdAt: now,
		updatedAt: now,
	}
}

// Getters
func (u *User) ID() string {
	return u.id
}

func (u *User) Email() valueobjects.Email {
	return u.email
}

func (u *User) Password() valueobjects.Password {
	return u.password
}

func (u *User) FirstName() string {
	if u.profile == nil {
		return ""
	}
	return u.profile.firstName
}

func (u *User) LastName() string {
	if u.profile == nil {
		return ""
	}
	return u.profile.lastName
}

func (u *User) Phone() string {
	if u.profile == nil {
		return ""
	}
	return u.profile.phone
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}

// Business methods
func (u *User) UpdateProfile(firstName, lastName, phone string) {
	if u.profile == nil {
		u.profile = &Profile{}
	}

	u.profile.firstName = firstName
	u.profile.lastName = lastName
	u.profile.phone = phone
	u.updatedAt = time.Now()
}

func (u *User) ChangePassword(newPassword valueobjects.Password) {
	u.password = newPassword
	u.updatedAt = time.Now()
}

func (u *User) ValidatePassword(password string) bool {
	return u.password.Verify(password)
}
