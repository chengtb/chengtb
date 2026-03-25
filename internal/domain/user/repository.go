package user

import "context"

// Repository defines the persistence contract for the User aggregate.
// Infrastructure implementations live in internal/infrastructure/persistence.
type Repository interface {
	// Save persists a new user or overwrites an existing one.
	Save(ctx context.Context, u *User) error

	// FindByID returns the user with the given identifier.
	// Returns errors.CodeNotFound when the user does not exist.
	FindByID(ctx context.Context, id string) (*User, error)

	// FindAll returns every user in the store.
	FindAll(ctx context.Context) ([]*User, error)

	// Delete removes the user identified by id.
	Delete(ctx context.Context, id string) error
}
