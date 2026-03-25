// Package persistence provides infrastructure-layer repository implementations.
package persistence

import (
	"context"
	"sync"

	domain "github.com/chengtb/chengtb/internal/domain/user"
	apperrors "github.com/chengtb/chengtb/pkg/errors"
)

// InMemoryUserRepository is a thread-safe, in-memory implementation of
// domain.Repository.  Replace it with a real database adapter in production.
type InMemoryUserRepository struct {
	mu    sync.RWMutex
	store map[string]*domain.User
}

// NewInMemoryUserRepository creates an empty repository.
func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{store: make(map[string]*domain.User)}
}

// Save stores or overwrites the user.
func (r *InMemoryUserRepository) Save(_ context.Context, u *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.store[u.ID()] = u
	return nil
}

// FindByID returns the user or a CodeNotFound error.
func (r *InMemoryUserRepository) FindByID(_ context.Context, id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.store[id]
	if !ok {
		return nil, apperrors.New(apperrors.CodeNotFound, "user not found")
	}
	return u, nil
}

// FindAll returns all stored users.
func (r *InMemoryUserRepository) FindAll(_ context.Context) ([]*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	users := make([]*domain.User, 0, len(r.store))
	for _, u := range r.store {
		users = append(users, u)
	}
	return users, nil
}

// Delete removes the user; returns CodeNotFound if it does not exist.
func (r *InMemoryUserRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.store[id]; !ok {
		return apperrors.New(apperrors.CodeNotFound, "user not found")
	}
	delete(r.store, id)
	return nil
}
