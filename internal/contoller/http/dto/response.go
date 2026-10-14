package dto

import (
	"main/internal/entity"

	"github.com/google/uuid"
)

type CreateResponseRequest struct {
	UserID  uuid.UUID             `json:"user_id" binding:"required"`
	Company string                `json:"company" binding:"required"`
	Vacancy string                `json:"vacancy" binding:"required"`
	Status  entity.ResponseStatus `json:"status,omitempty"`
}

type ResponseResponse struct {
	ID      uuid.UUID             `json:"id"`
	UserID  uuid.UUID             `json:"user_id"`
	Company string                `json:"company"`
	Vacancy string                `json:"vacancy"`
	Status  entity.ResponseStatus `json:"status"`
}
