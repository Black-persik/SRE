package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"main/internal/contoller/http/dto"
	"main/internal/contoller/http/middleware"
	"main/internal/entity"
	"main/internal/error"
	"net/http"

	"github.com/google/uuid"
)

type UserService interface {
	Create(ctx context.Context, user entity.CreateUser) (entity.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (entity.User, error)
	GetByUsername(ctx context.Context, username string) (entity.User, error)
	GetByEmail(ctx context.Context, email string) (entity.User, error)
	UpdateEmail(ctx context.Context, id uuid.UUID, email string) error
}

type UserHandler struct {
	users UserService
}

func NewUserHandler(users UserService) *UserHandler {
	return &UserHandler{users: users}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.Write(w, apperror.ErrInvalidJSON)
		return
	}

	if req.Username == "" || req.Email == "" {
		apperror.Write(w, apperror.ErrUsernameAndEmailRequired)
		return
	}

	user, err := h.users.Create(r.Context(), entity.CreateUser{
		Username: req.Username,
		Email:    req.Email,
	})
	if errors.Is(err, apperror.ErrUsernameTaken) {
		apperror.Write(w, apperror.ErrUsernameAlreadyExists)
		return
	}
	if errors.Is(err, apperror.ErrEmailTaken) {
		apperror.Write(w, apperror.ErrEmailAlreadyExists)
		return
	}
	if err != nil {
		middleware.SetError(r, err)
		apperror.Write(w, apperror.ErrInternal)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		apperror.Write(w, apperror.ErrInvalidID)
		return
	}
	user, err := h.users.GetByID(r.Context(), id)
	if errors.Is(err, apperror.ErrNotFound) {
		apperror.Write(w, apperror.ErrUserMissing)
		return
	}
	if err != nil {
		middleware.SetError(r, err)
		apperror.Write(w, apperror.ErrInternal)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) GetUserByUsername(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if username == "" {
		apperror.Write(w, apperror.ErrUsernameRequired)
		return
	}
	user, err := h.users.GetByUsername(r.Context(), username)
	if errors.Is(err, apperror.ErrNotFound) {
		apperror.Write(w, apperror.ErrUserMissing)
		return
	}
	if err != nil {
		middleware.SetError(r, err)
		apperror.Write(w, apperror.ErrInternal)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) UpdateUserEmail(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		apperror.Write(w, apperror.ErrInvalidID)
		return
	}
	email := r.PathValue("email")
	if email == "" {
		apperror.Write(w, apperror.ErrEmailRequired)
		return
	}

	if err = h.users.UpdateEmail(r.Context(), id, email); err != nil {
		if errors.Is(err, apperror.ErrEmailTaken) {
			apperror.Write(w, apperror.ErrEmailAlreadyExists)
			return
		}
		middleware.SetError(r, err)
		apperror.Write(w, apperror.ErrInternal)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "email updated successfully"})
}
