package user

import (
	"context"

	domain "github.com/chengtb/chengtb/internal/domain/user"
	apperrors "github.com/chengtb/chengtb/pkg/errors"
	"github.com/chengtb/chengtb/pkg/logger"
	"go.uber.org/zap"
)

// Service orchestrates use-cases by coordinating the domain aggregate,
// domain service and repository.  It sits at the boundary between the
// application layer and the domain layer.
type Service struct {
	repo          domain.Repository
	domainService *domain.DomainService
}

// NewService constructs the application service.
func NewService(repo domain.Repository) *Service {
	return &Service{
		repo:          repo,
		domainService: domain.NewDomainService(repo),
	}
}

// CreateUser handles the CreateUserCommand.
func (s *Service) CreateUser(ctx context.Context, cmd CreateUserCommand) (*UserDTO, error) {
	if err := s.domainService.EnsureEmailUnique(ctx, cmd.Email, ""); err != nil {
		return nil, err
	}
	u, err := domain.New(cmd.ID, cmd.Name, cmd.Email)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, u); err != nil {
		return nil, err
	}
	logger.L().Info("user created", zap.String("id", u.ID()))
	return toDTO(u), nil
}

// GetUser handles the GetUserQuery.
func (s *Service) GetUser(ctx context.Context, q GetUserQuery) (*UserDTO, error) {
	if q.ID == "" {
		return nil, apperrors.New(apperrors.CodeInvalidArg, "user id is required")
	}
	u, err := s.repo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, err
	}
	return toDTO(u), nil
}

// ListUsers handles the ListUsersQuery.
func (s *Service) ListUsers(ctx context.Context, _ ListUsersQuery) ([]*UserDTO, error) {
	users, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	dtos := make([]*UserDTO, 0, len(users))
	for _, u := range users {
		dtos = append(dtos, toDTO(u))
	}
	return dtos, nil
}

// UpdateUser handles the UpdateUserCommand.
func (s *Service) UpdateUser(ctx context.Context, cmd UpdateUserCommand) (*UserDTO, error) {
	u, err := s.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}
	if cmd.Email != "" && cmd.Email != u.Email() {
		if err := s.domainService.EnsureEmailUnique(ctx, cmd.Email, cmd.ID); err != nil {
			return nil, err
		}
		if err := u.ChangeEmail(cmd.Email); err != nil {
			return nil, err
		}
	}
	if cmd.Name != "" {
		if err := u.Rename(cmd.Name); err != nil {
			return nil, err
		}
	}
	if err := s.repo.Save(ctx, u); err != nil {
		return nil, err
	}
	logger.L().Info("user updated", zap.String("id", u.ID()))
	return toDTO(u), nil
}

// DeleteUser handles the DeleteUserCommand.
func (s *Service) DeleteUser(ctx context.Context, cmd DeleteUserCommand) error {
	if cmd.ID == "" {
		return apperrors.New(apperrors.CodeInvalidArg, "user id is required")
	}
	if err := s.repo.Delete(ctx, cmd.ID); err != nil {
		return err
	}
	logger.L().Info("user deleted", zap.String("id", cmd.ID))
	return nil
}

// toDTO converts a domain User to a UserDTO.
func toDTO(u *domain.User) *UserDTO {
	return &UserDTO{
		ID:        u.ID(),
		Name:      u.Name(),
		Email:     u.Email(),
		CreatedAt: u.CreatedAt(),
		UpdatedAt: u.UpdatedAt(),
	}
}
