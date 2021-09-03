package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config represents service configuration for dp-observation-api
type Config struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	HttpWriteTimeout           time.Duration `envconfig:"HTTP_WRITE_TIMEOUT"`
	CantabularRequestTimeout   time.Duration `envconfig:"CANTABULAR_REQUEST_TIMEOUT`
	ServiceAuthToken           string        `envconfig:"SERVICE_AUTH_TOKEN"`
	DatasetAPIURL              string        `envconfig:"DATASET_API_URL"`
	ObservationAPIURL          string        `envconfig:"OBSERVATION_API_URL"`
	ZebedeeURL                 string        `envconfig:"ZEBEDEE_URL"`
	CantabularURL              string        `envconfig:"CANTABULAR_URL"`
	CantabularExtURL           string        `envconfig:"CANTABULAR_EXT_API_URL"`
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
		HttpWriteTimeout:           60 * time.Second,
		CantabularRequestTimeout:   10 * time.Second,
		ServiceAuthToken:           "",
		DatasetAPIURL:              "http://localhost:22000",
		ObservationAPIURL:          "http://localhost:24500",
		ZebedeeURL:                 "http://localhost:8082",
		CantabularURL:              "localhost:8491",
		CantabularExtURL:           "http://localhost:8492",
		DefaultObservationLimit:    10000,
		EnablePrivateEndpoints:     false,
		GracefulShutdownTimeout:    5 * time.Second,
		HealthCheckInterval:        30 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
	}

	return cfg, envconfig.Process("", cfg)
}
