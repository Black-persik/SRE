// Package metrics declares the application's Prometheus metrics in one place.
// Label values come only from the constants below, never from user input,
// so the number of time series stays bounded.
package metrics

import (
	"errors"
	"main/internal/entity"
	"main/internal/error"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Operation string

const (
	OpCreateUser     Operation = "create_user"
	OpUpdateEmail    Operation = "update_email"
	OpCreateResponse Operation = "create_response"
	OpUpdateStatus   Operation = "update_status"
	OpDeleteResponse Operation = "delete_response"
)

const (
	resultSuccess = "success"
	// The request was valid but the domain refused it (taken username, unknown user, ...).
	resultRejected = "rejected"
	// The operation failed for an internal reason (database down, bug, ...).
	resultError = "error"
)

var (
	businessOperations = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "business_operations_total",
		Help: "Business operations by type and result.",
	}, []string{"operation", "result"})

	userRegistrations = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "user_registrations_total",
		Help: "User registration attempts by result.",
	}, []string{"result"})

	responsesCreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "responses_created_total",
		Help: "Job responses created.",
	})

	responseStatusChanges = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "response_status_changes_total",
		Help: "Job response status changes by new status.",
	}, []string{"status"})
)

// ObserveOperation counts a finished business operation; err is what the service returned.
func ObserveOperation(op Operation, err error) {
	businessOperations.WithLabelValues(string(op), result(err)).Inc()
}

func ObserveUserRegistration(err error) {
	switch {
	case err == nil:
		userRegistrations.WithLabelValues(resultSuccess).Inc()
	case errors.Is(err, apperror.ErrUsernameTaken):
		userRegistrations.WithLabelValues("username_taken").Inc()
	case errors.Is(err, apperror.ErrEmailTaken):
		userRegistrations.WithLabelValues("email_taken").Inc()
	default:
		userRegistrations.WithLabelValues(resultError).Inc()
	}
}

func ResponseCreated() {
	responsesCreated.Inc()
}

// ResponseStatusChanged expects a status that already passed entity.ResponseStatus.Valid.
func ResponseStatusChanged(status entity.ResponseStatus) {
	responseStatusChanges.WithLabelValues(string(status)).Inc()
}

func result(err error) string {
	switch {
	case err == nil:
		return resultSuccess
	case errors.Is(err, apperror.ErrUsernameTaken),
		errors.Is(err, apperror.ErrEmailTaken),
		errors.Is(err, apperror.ErrUserNotFound),
		errors.Is(err, apperror.ErrNotFound),
		errors.Is(err, apperror.ErrInvalidStatus):
		return resultRejected
	default:
		return resultError
	}
}
