package service

import (
	"context"
	"errors"
	"testing"

	"main/internal/entity"
	"main/internal/error"

	"github.com/google/uuid"
)

type userRepoStub struct {
	byUsername entity.User
	byEmail    entity.User
	err        error
}

func (s userRepoStub) Create(context.Context, entity.CreateUser) (entity.User, error) {
	return entity.User{}, s.err
}

func (s userRepoStub) GetByID(context.Context, uuid.UUID) (entity.User, error) {
	return entity.User{}, apperror.ErrNotFound
}

func (s userRepoStub) GetByUsername(context.Context, string) (entity.User, error) {
	if s.byUsername.Username != "" {
		return s.byUsername, nil
	}
	return entity.User{}, apperror.ErrNotFound
}

func (s userRepoStub) GetByEmail(context.Context, string) (entity.User, error) {
	if s.byEmail.Email != "" {
		return s.byEmail, nil
	}
	return entity.User{}, apperror.ErrNotFound
}

func (s userRepoStub) UpdateEmail(context.Context, uuid.UUID, string) error {
	return nil
}

func TestCreateRejectsExistingUsername(t *testing.T) {
	svc := NewUserService(userRepoStub{byUsername: entity.User{Username: "ivan"}})

	_, err := svc.Create(context.Background(), entity.CreateUser{Username: "ivan", Email: "a@b.c"})
	if !errors.Is(err, apperror.ErrUsernameTaken) {
		t.Fatalf("err = %v", err)
	}
}

func TestCreateRejectsExistingEmail(t *testing.T) {
	svc := NewUserService(userRepoStub{byEmail: entity.User{Email: "a@b.c"}})

	_, err := svc.Create(context.Background(), entity.CreateUser{Username: "ivan", Email: "a@b.c"})
	if !errors.Is(err, apperror.ErrEmailTaken) {
		t.Fatalf("err = %v", err)
	}
}
