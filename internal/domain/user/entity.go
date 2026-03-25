// Package user contains the User aggregate root and related domain objects.
package user

import (
	"time"

	apperrors "github.com/chengtb/chengtb/pkg/errors"
)

// User is the aggregate root for the User domain.
type User struct {
	id        string
	name      string
	email     string
	createdAt time.Time
	updatedAt time.Time
}

// New creates a new User entity, validating invariants.
func New(id, name, email string) (*User, error) {
	if id == "" {
		return nil, apperrors.New(apperrors.CodeInvalidArg, "user id is required")
	}
	if name == "" {
		return nil, apperrors.New(apperrors.CodeInvalidArg, "user name is required")
	}
	if email == "" {
		return nil, apperrors.New(apperrors.CodeInvalidArg, "user email is required")
	}
	now := time.Now()
	return &User{
		id:        id,
		name:      name,
		email:     email,
		createdAt: now,
		updatedAt: now,
	}, nil
}

// Reconstitute rebuilds a User from its persisted state (bypasses invariant
// checks that only apply on creation).
func Reconstitute(id, name, email string, createdAt, updatedAt time.Time) *User {
	return &User{
		id:        id,
		name:      name,
		email:     email,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

// ID returns the user's identifier.
func (u *User) ID() string { return u.id }

// Name returns the user's full name.
func (u *User) Name() string { return u.name }

// Email returns the user's email address.
func (u *User) Email() string { return u.email }

// CreatedAt returns the creation timestamp.
func (u *User) CreatedAt() time.Time { return u.createdAt }

// UpdatedAt returns the last-update timestamp.
func (u *User) UpdatedAt() time.Time { return u.updatedAt }

// Rename updates the user's name, enforcing invariants.
func (u *User) Rename(name string) error {
	if name == "" {
		return apperrors.New(apperrors.CodeInvalidArg, "user name cannot be empty")
	}
	u.name = name
	u.updatedAt = time.Now()
	return nil
}

// ChangeEmail updates the user's email address, enforcing invariants.
func (u *User) ChangeEmail(email string) error {
	if email == "" {
		return apperrors.New(apperrors.CodeInvalidArg, "user email cannot be empty")
	}
	u.email = email
	u.updatedAt = time.Now()
	return nil
}
