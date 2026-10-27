package prober

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type step string

// Steps of the scenario in the order they run.
const (
	stepCreateUser     step = "create_user"
	stepGetUser        step = "get_user"
	stepCreateResponse step = "create_response"
	stepUpdateStatus   step = "update_status"
	stepGetResponse    step = "get_response"
	stepDeleteResponse step = "delete_response"
)

const (
	resultSuccess = "success"
	resultFail    = "fail"
)

var (
	stepsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "prober_steps_total",
		Help: "Prober scenario steps by step and result.",
	}, []string{"step", "result"})

	stepDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "prober_step_duration_seconds",
		Help:    "Duration of prober scenario steps, failed ones included.",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
	}, []string{"step"})

	scenariosTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "prober_scenarios_total",
		Help: "Full prober scenario runs by result.",
	}, []string{"result"})

	lastSuccess = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "prober_last_success_timestamp_seconds",
		Help: "Unix time of the last fully successful scenario run.",
	})
)

func init() {
	// Pre-create series so rate() works before the first failure happens.
	for _, s := range []step{stepCreateUser, stepGetUser, stepCreateResponse, stepUpdateStatus, stepGetResponse, stepDeleteResponse} {
		stepsTotal.WithLabelValues(string(s), resultSuccess)
		stepsTotal.WithLabelValues(string(s), resultFail)
	}
	scenariosTotal.WithLabelValues(resultSuccess)
	scenariosTotal.WithLabelValues(resultFail)
}
