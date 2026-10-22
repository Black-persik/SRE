package entity

import "github.com/google/uuid"

type ResponseStatus string

const (
	WaitingStatus  ResponseStatus = "waiting"
	AcceptedStatus ResponseStatus = "accepted"
	RejectedStatus ResponseStatus = "rejected"
)

func (s ResponseStatus) Valid() bool {
	switch s {
	case WaitingStatus, AcceptedStatus, RejectedStatus:
		return true
	}
	return false
}

type Response struct {
	ID      uuid.UUID
	UserID  uuid.UUID
	Company string
	Vacancy string
	Status  ResponseStatus
}

type CreateResponse struct {
	UserID  uuid.UUID
	Company string
	Vacancy string
	Status  ResponseStatus
}