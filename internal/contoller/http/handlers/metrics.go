package handlers

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewMetricsHandler() http.Handler {
	return promhttp.Handler()
}