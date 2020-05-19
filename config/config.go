package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config represents service configuration for dp-observation-api
type Config struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	ServiceAuthToken           string        `envconfig:"SERVICE_AUTH_TOKEN"`
	DatasetAPIURL              string        `envconfig:"DATASET_API_URL"`
	ObservationAPIURL          string        `envconfig:"OBSERVATION_API_URL"`
	ZebedeeURL                 string        `envconfig:"ZEBEDEE_URL"`
	DefaultObservationLimit    int           `envconfig:"DEFAULT_OBSERVATION_LIMIT"`
	EnablePrivateEndpoints     bool          `envconfig:"ENABLE_PRIVATE_ENDPOINTS"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
}

var cfg *Config

// Get returns the default config with any modifications through environment
// variables
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg := &Config{
		BindAddr:                   ":24500",
		ServiceAuthToken:           "",
		DatasetAPIURL:              "http://localhost:22000",
		ObservationAPIURL:          "http://localhost:24500",
		ZebedeeURL:                 "http://localhost:8082",
		DefaultObservationLimit:    10000,
		EnablePrivateEndpoints:     false,
		GracefulShutdownTimeout:    5 * time.Second,
		HealthCheckInterval:        30 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
	}

	return cfg, envconfig.Process("", cfg)
}
