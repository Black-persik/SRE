package service

import (
	"context"
	"errors"
	"testing"

	"main/internal/entity"
	"main/internal/error"

	"github.com/google/uuid"
)

type responseRepoStub struct{}

func (responseRepoStub) Create(context.Context, entity.CreateResponse) (entity.Response, error) {
	return entity.Response{}, nil
}

func (responseRepoStub) GetByID(context.Context, uuid.UUID) (entity.Response, error) {
	return entity.Response{}, nil
}

func (responseRepoStub) GetByUserID(context.Context, uuid.UUID) ([]entity.Response, error) {
	return nil, nil
}

func (responseRepoStub) UpdateStatus(context.Context, uuid.UUID, entity.ResponseStatus) error {
	return nil
}

func (responseRepoStub) Delete(context.Context, uuid.UUID) error {
	return nil
}

type missingUserRepo struct{}

func (missingUserRepo) GetByID(context.Context, uuid.UUID) (entity.User, error) {
	return entity.User{}, apperror.ErrNotFound
}

func TestCreateResponseRejectsMissingUser(t *testing.T) {
	svc := NewResponseService(responseRepoStub{}, missingUserRepo{})

	_, err := svc.Create(context.Background(), entity.CreateResponse{
		UserID:  uuid.New(),
		Company: "acme",
		Vacancy: "sre",
		Status:  entity.WaitingStatus,
	})
	if !errors.Is(err, apperror.ErrUserNotFound) {
		t.Fatalf("err = %v", err)
	}
}
