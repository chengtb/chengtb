package user

import (
	"context"

	apperrors "github.com/chengtb/chengtb/pkg/errors"
)

// DomainService encapsulates business rules that do not naturally fit
// inside a single aggregate (e.g. uniqueness invariants).
type DomainService struct {
	repo Repository
}

// NewDomainService constructs the domain service with its repository dependency.
func NewDomainService(repo Repository) *DomainService {
	return &DomainService{repo: repo}
}

// EnsureEmailUnique returns an error when another user already uses the email.
func (s *DomainService) EnsureEmailUnique(ctx context.Context, email, excludeID string) error {
	all, err := s.repo.FindAll(ctx)
	if err != nil {
		return err
	}
	for _, u := range all {
		if u.Email() == email && u.ID() != excludeID {
			return apperrors.New(apperrors.CodeConflict, "email already in use")
		}
	}
	return nil
}
