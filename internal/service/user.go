package service

import (
	"context"
	"errors"
	"fmt"
	"main/internal/entity"
	"main/internal/error"
	"main/internal/metrics"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user entity.CreateUser) (entity.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (entity.User, error)
	GetByUsername(ctx context.Context, username string) (entity.User, error)
	GetByEmail(ctx context.Context, email string) (entity.User, error)
	UpdateEmail(ctx context.Context, id uuid.UUID, email string) error
}

type UserService struct {
	users UserRepository
}

func NewUserService(users UserRepository) *UserService {
	return &UserService{users: users}
}

func (s *UserService) Create(ctx context.Context, user entity.CreateUser) (_ entity.User, err error) {
	defer func() {
		metrics.ObserveUserRegistration(err)
		metrics.ObserveOperation(metrics.OpCreateUser, err)
	}()

	_, err = s.users.GetByUsername(ctx, user.Username)
	if err == nil {
		return entity.User{}, apperror.ErrUsernameTaken
	}
	if !errors.Is(err, apperror.ErrNotFound) {
		return entity.User{}, fmt.Errorf("UserService.Create: get by username: %w", err)
	}

	_, err = s.users.GetByEmail(ctx, user.Email)
	if err == nil {
		return entity.User{}, apperror.ErrEmailTaken
	}
	if !errors.Is(err, apperror.ErrNotFound) {
		return entity.User{}, fmt.Errorf("UserService.Create: get by email: %w", err)
	}

	createdUser, err := s.users.Create(ctx, user)
	if err != nil {
		if errors.Is(err, apperror.ErrUsernameTaken) || errors.Is(err, apperror.ErrEmailTaken) {
			return entity.User{}, err
		}
		return entity.User{}, fmt.Errorf("UserService.Create: create user: %w", err)
	}
	return createdUser, nil
}

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (entity.User, error) {
	user, err := s.users.GetByID(ctx, id)
	if err != nil {
		return entity.User{}, fmt.Errorf("UserService.GetByID: get user by id: %w", err)
	}
	return user, nil
}

func (s *UserService) GetByUsername(ctx context.Context, username string) (entity.User, error) {
	user, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		return entity.User{}, fmt.Errorf("UserService.GetByUsername: get user by username: %w", err)
	}
	return user, nil
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (entity.User, error) {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return entity.User{}, fmt.Errorf("UserService.GetByEmail: get user by email: %w", err)
	}
	return user, nil
}

func (s *UserService) UpdateEmail(ctx context.Context, id uuid.UUID, email string) (err error) {
	defer func() { metrics.ObserveOperation(metrics.OpUpdateEmail, err) }()

	existing, err := s.users.GetByEmail(ctx, email)
	if err == nil && existing.ID != id {
		return apperror.ErrEmailTaken
	}
	if err != nil && !errors.Is(err, apperror.ErrNotFound) {
		return fmt.Errorf("UserService.UpdateEmail: get by email: %w", err)
	}

	if err = s.users.UpdateEmail(ctx, id, email); err != nil {
		if errors.Is(err, apperror.ErrEmailTaken) {
			return apperror.ErrEmailTaken
		}
		return fmt.Errorf("UserService.UpdateEmail: update email: %w", err)
	}
	return nil
}
