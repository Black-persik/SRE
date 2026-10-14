package service

import (
	"context"
	"errors"
	"fmt"
	"main/internal/entity"
	"main/internal/error"

	"github.com/google/uuid"
)

type ResponseRepository interface {
	Create(ctx context.Context, response entity.CreateResponse) (entity.Response, error)
	GetByID(ctx context.Context, id uuid.UUID) (entity.Response, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Response, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ResponseStatus) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type userByID interface {
	GetByID(ctx context.Context, id uuid.UUID) (entity.User, error)
}

type ResponseService struct {
	responses ResponseRepository
	users     userByID
}

func NewResponseService(responses ResponseRepository, users userByID) *ResponseService {
	return &ResponseService{responses: responses, users: users}
}

func (s *ResponseService) Create(ctx context.Context, response entity.CreateResponse) (entity.Response, error) {
	_, err := s.users.GetByID(ctx, response.UserID)
	if errors.Is(err, apperror.ErrNotFound) {
		return entity.Response{}, apperror.ErrUserNotFound
	}
	if err != nil {
		return entity.Response{}, fmt.Errorf("ResponseService.Create: get user: %w", err)
	}

	createdResponse, err := s.responses.Create(ctx, response)
	if errors.Is(err, apperror.ErrUserNotFound) {
		return entity.Response{}, apperror.ErrUserNotFound
	}
	if err != nil {
		return entity.Response{}, fmt.Errorf("ResponseService.Create: create response: %w", err)
	}
	return createdResponse, nil
}

func (s *ResponseService) GetByID(ctx context.Context, id uuid.UUID) (entity.Response, error) {
	response, err := s.responses.GetByID(ctx, id)
	if err != nil {
		return entity.Response{}, fmt.Errorf("ResponseService.GetByID: get response by id: %w", err)
	}
	return response, nil
}

func (s *ResponseService) GetByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Response, error) {
	responses, err := s.responses.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ResponseService.GetByUserID: get responses by user id: %w", err)
	}
	return responses, nil
}

func (s *ResponseService) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ResponseStatus) error {
	err := s.responses.UpdateStatus(ctx, id, status)
	if err != nil {
		return fmt.Errorf("ResponseService.UpdateStatus: update status: %w", err)
	}
	return nil
}

func (s *ResponseService) Delete(ctx context.Context, id uuid.UUID) error {
	err := s.responses.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("ResponseService.Delete: delete response: %w", err)
	}
	return nil
}
