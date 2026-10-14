package entity

import "github.com/google/uuid"

type User struct {
	ID       uuid.UUID
	Username string
	Email    string
}

type CreateUser struct {
	Username string
	Email    string
}
