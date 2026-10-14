package config

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type (
	Config struct {
		App            AppConfig      `yaml:"app"`
		Database       DatabaseConfig `yaml:"db"`
		DatabaseSecret DatabaseSecret
	}

	DatabaseConfig struct {
		ConnectTimeout  time.Duration `yaml:"connect_timeout"`
		MaxIdleConns    int           `yaml:"max_idle_conns"`
		MaxOpenConns    int           `yaml:"max_open_conns"`
		ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
		ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
	}

	DatabaseSecret struct {
		User     string `env:"DB_USER"`
		Password string `env:"DB_PASSWORD"`
		Host     string `env:"DB_HOST"`
		Port     int    `env:"DB_PORT"`
		Name     string `env:"DB_NAME"`
	}

	AppConfig struct {
		Port        int    `yaml:"port"`
		Environment string `env:"ENVIRONMENT"`
	}
)

func NewConfig(ctx context.Context) (*Config, error) {
	cfg := &Config{}

	if err := cfg.readSecrets(); err != nil {
		return nil, fmt.Errorf("cfg.readSecrets: %w", err)
	}

	if err := cfg.readConfig(); err != nil {
		return nil, fmt.Errorf("cfg.readConfig: %w", err)
	}

	return cfg, nil
}

//go:embed config_dev.yaml config_prod.yaml config_staging.yaml
var configFS embed.FS

func (c *Config) readConfig() error {
	envName := c.App.Environment
	if envName == "" {
		envName = "dev"
	}

	fileName := fmt.Sprintf("config_%s.yaml", envName)

	b, err := configFS.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("failed to read config file %q: %w", fileName, err)
	}

	if err := yaml.Unmarshal(b, c); err != nil {
		return fmt.Errorf("failed to unmarshal config file %q: %w", fileName, err)
	}

	return nil
}

func (c *Config) readSecrets() error {
	if err := godotenv.Load(); err != nil && ! errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("godotenv.Load: %w", err)
	}

	if err := env.Parse(c); err != nil {
		return fmt.Errorf("env.Parse: %w", err)
	}
	return nil
}
