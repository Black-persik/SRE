package entity

import "github.com/google/uuid"

type ResponseStatus string

const (
	WaitingStatus  ResponseStatus = "waiting"
	AcceptedStatus ResponseStatus = "accepted"
	RejectedStatus ResponseStatus = "rejected"
)

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