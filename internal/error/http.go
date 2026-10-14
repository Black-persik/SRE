package apperror

import "net/http"

type HTTPError struct {
	Status  int
	Message string
}

func (e HTTPError) Error() string {
	return e.Message
}

func Write(w http.ResponseWriter, err HTTPError) {
	http.Error(w, err.Message, err.Status)
}

var (
	ErrInvalidJSON              = HTTPError{Status: http.StatusBadRequest, Message: "invalid json"}
	ErrInvalidID                = HTTPError{Status: http.StatusBadRequest, Message: "invalid id"}
	ErrInvalidUserID            = HTTPError{Status: http.StatusBadRequest, Message: "invalid user id"}
	ErrUsernameAndEmailRequired = HTTPError{Status: http.StatusBadRequest, Message: "username and email are required"}
	ErrUsernameRequired         = HTTPError{Status: http.StatusBadRequest, Message: "username is required"}
	ErrEmailRequired            = HTTPError{Status: http.StatusBadRequest, Message: "email is required"}
	ErrResponseFieldsRequired   = HTTPError{Status: http.StatusBadRequest, Message: "user_id, company, and vacancy are required"}
	ErrStatusRequired           = HTTPError{Status: http.StatusBadRequest, Message: "status is required"}
	ErrUsernameAlreadyExists    = HTTPError{Status: http.StatusConflict, Message: ErrUsernameTaken.Error()}
	ErrEmailAlreadyExists       = HTTPError{Status: http.StatusConflict, Message: ErrEmailTaken.Error()}
	ErrUserMissing              = HTTPError{Status: http.StatusNotFound, Message: ErrUserNotFound.Error()}
	ErrInternal                 = HTTPError{Status: http.StatusInternalServerError, Message: "internal error"}
)
