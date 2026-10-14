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

type ResponseService interface {
	Create(ctx context.Context, response entity.CreateResponse) (entity.Response, error)
	GetByID(ctx context.Context, id uuid.UUID) (entity.Response, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]entity.Response, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.ResponseStatus) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type ResponseHandler struct {
	responses ResponseService
}

func NewResponseHandler(responses ResponseService) *ResponseHandler {
	return &ResponseHandler{responses: responses}
}

func (h *ResponseHandler) CreateResponse(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateResponseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apperror.Write(w, apperror.ErrInvalidJSON)
		return
	}

	if req.UserID == uuid.Nil || req.Company == "" || req.Vacancy == "" {
		apperror.Write(w, apperror.ErrResponseFieldsRequired)
		return
	}

	if req.Status == "" {
		req.Status = entity.WaitingStatus
	}

	resp, err := h.responses.Create(r.Context(), entity.CreateResponse{
		UserID:  req.UserID,
		Company: req.Company,
		Vacancy: req.Vacancy,
		Status:  req.Status,
	})
	if errors.Is(err, apperror.ErrUserNotFound) {
		apperror.Write(w, apperror.ErrUserMissing)
		return
	}
	if err != nil {
		middleware.SetError(r, err)
		apperror.Write(w, apperror.ErrInternal)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *ResponseHandler) GetResponseByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		apperror.Write(w, apperror.ErrInvalidID)
		return
	}

	resp, err := h.responses.GetByID(r.Context(), id)
	if err != nil {
		middleware.SetError(r, err)
		apperror.Write(w, apperror.ErrInternal)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *ResponseHandler) GetResponsesByUserID(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(r.PathValue("user_id"))
	if err != nil {
		apperror.Write(w, apperror.ErrInvalidUserID)
		return
	}

	resp, err := h.responses.GetByUserID(r.Context(), userID)
	if err != nil {
		middleware.SetError(r, err)
		apperror.Write(w, apperror.ErrInternal)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *ResponseHandler) UpdateResponseStatus(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		apperror.Write(w, apperror.ErrInvalidID)
		return
	}
	newStatus := entity.ResponseStatus(r.PathValue("status"))
	if newStatus == "" {
		apperror.Write(w, apperror.ErrStatusRequired)
		return
	}

	if err = h.responses.UpdateStatus(r.Context(), id, newStatus); err != nil {
		middleware.SetError(r, err)
		apperror.Write(w, apperror.ErrInternal)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "status updated successfully"})
}

func (h *ResponseHandler) DeleteResponse(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		apperror.Write(w, apperror.ErrInvalidID)
		return
	}

	if err = h.responses.Delete(r.Context(), id); err != nil {
		middleware.SetError(r, err)
		apperror.Write(w, apperror.ErrInternal)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "response deleted successfully"})

}
