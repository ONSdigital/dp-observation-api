package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config represents service configuration for dp-observation-api
type Config struct {
	BindAddr                     string        `envconfig:"BIND_ADDR"`
	HTTPWriteTimeout             time.Duration `envconfig:"HTTP_WRITE_TIMEOUT"`
	CantabularRequestTimeout     time.Duration `envconfig:"CANTABULAR_REQUEST_TIMEOUT"`
	ServiceAuthToken             string        `envconfig:"SERVICE_AUTH_TOKEN"          json:"-"`
	CodeListAPIURL               string        `envconfig:"CODE_LIST_API_URL"`
	DatasetAPIURL                string        `envconfig:"DATASET_API_URL"`
	ObservationAPIURL            string        `envconfig:"OBSERVATION_API_URL"`
	ZebedeeURL                   string        `envconfig:"ZEBEDEE_URL"`
	CantabularURL                string        `envconfig:"CANTABULAR_URL"`
	CantabularExtURL             string        `envconfig:"CANTABULAR_EXT_API_URL"`
	CantabularHealthcheckEnabled bool          `envconfig:"CANTABULAR_HEALTHCHECK_ENABLED"`
	DefaultObservationLimit      int           `envconfig:"DEFAULT_OBSERVATION_LIMIT"`
	EnablePrivateEndpoints       bool          `envconfig:"ENABLE_PRIVATE_ENDPOINTS"`
	EnableURLRewriting           bool          `envconfig:"ENABLE_URL_REWRITING"`
	GracefulShutdownTimeout      time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval          time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout   time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
}

var cfg *Config

// Get returns the default config with any modifications through environment
// variables
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg := &Config{
		BindAddr:                     ":24500",
		HTTPWriteTimeout:             60 * time.Second,
		CantabularRequestTimeout:     10 * time.Second,
		ServiceAuthToken:             "",
		CodeListAPIURL:               "http://localhost:22400",
		DatasetAPIURL:                "http://localhost:22000",
		ObservationAPIURL:            "http://localhost:24500",
		ZebedeeURL:                   "http://localhost:8082",
		CantabularURL:                "localhost:8491",
		CantabularExtURL:             "http://localhost:8492",
		CantabularHealthcheckEnabled: false,
		DefaultObservationLimit:      10000,
		EnablePrivateEndpoints:       false,
		EnableURLRewriting:           false,
		GracefulShutdownTimeout:      5 * time.Second,
		HealthCheckInterval:          30 * time.Second,
		HealthCheckCriticalTimeout:   90 * time.Second,
	}

	return cfg, envconfig.Process("", cfg)
}
