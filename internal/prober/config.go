package prober

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	// Base URL of the application, e.g. http://app inside the cluster.
	TargetURL string        `env:"PROBER_TARGET_URL" envDefault:"http://localhost:8080"`
	Interval  time.Duration `env:"PROBER_INTERVAL" envDefault:"15s"`
	// Timeout for a single HTTP request, not for the whole scenario.
	Timeout     time.Duration `env:"PROBER_TIMEOUT" envDefault:"5s"`
	MetricsPort int           `env:"PROBER_METRICS_PORT" envDefault:"9090"`
}

func NewConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("env.Parse: %w", err)
	}
	return cfg, nil
}
