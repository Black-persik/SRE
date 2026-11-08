package prober

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// check is an independent part of a prober run; one failing check does not stop the others.
type check string

const (
	checkRegistration check = "registration"
	checkResponse     check = "response"
	checkStatus       check = "status"
)

type step string

// Steps of the checks in the order they run.
const (
	stepCreateUser     step = "create_user"
	stepGetUser        step = "get_user"
	stepCreateResponse step = "create_response"
	stepDeleteResponse step = "delete_response"
	stepUpdateStatus   step = "update_status"
	stepGetResponse    step = "get_response"
)

const (
	resultSuccess = "success"
	resultFail    = "fail"
)

var (
	stepsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "prober_steps_total",
		Help: "Prober check steps by step and result.",
	}, []string{"step", "result"})

	stepDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "prober_step_duration_seconds",
		Help:    "Duration of prober check steps, failed ones included.",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
	}, []string{"step"})

	checksTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "prober_checks_total",
		Help: "Prober checks by check and result.",
	}, []string{"check", "result"})

	scenariosTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "prober_scenarios_total",
		Help: "Prober runs by result; a run succeeds only if every check succeeds.",
	}, []string{"result"})

	lastSuccess = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "prober_last_success_timestamp_seconds",
		Help: "Unix time of the last fully successful prober run.",
	})
)

func init() {
	// Pre-create series so rate() works before the first failure happens.
	for _, s := range []step{stepCreateUser, stepGetUser, stepCreateResponse, stepDeleteResponse, stepUpdateStatus, stepGetResponse} {
		stepsTotal.WithLabelValues(string(s), resultSuccess)
		stepsTotal.WithLabelValues(string(s), resultFail)
	}
	for _, c := range []check{checkRegistration, checkResponse, checkStatus} {
		checksTotal.WithLabelValues(string(c), resultSuccess)
		checksTotal.WithLabelValues(string(c), resultFail)
	}
	scenariosTotal.WithLabelValues(resultSuccess)
	scenariosTotal.WithLabelValues(resultFail)
}
