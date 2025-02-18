package config

import (
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfig(t *testing.T) {
	os.Clearenv()

	Convey("Given an environment with no environment variables set", t, func() {
		cfg, err := Get()

		Convey("When the config values are retrieved", func() {
			Convey("Then there should be no error returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the values should be set to the expected defaults", func() {
				So(cfg, ShouldResemble, &Config{
					BindAddr:                   ":24500",
					HTTPWriteTimeout:           60 * time.Second,
					CantabularRequestTimeout:   10 * time.Second,
					ServiceAuthToken:           "",
					CodeListAPIURL:             "http://localhost:22400",
					DatasetAPIURL:              "http://localhost:22000",
					ObservationAPIURL:          "http://localhost:24500",
					ZebedeeURL:                 "http://localhost:8082",
					CantabularURL:              "localhost:8491",
					CantabularExtURL:           "http://localhost:8492",
					EnablePrivateEndpoints:     false,
					EnableURLRewriting:         false,
					DefaultObservationLimit:    10000,
					GracefulShutdownTimeout:    5 * time.Second,
					HealthCheckInterval:        30 * time.Second,
					HealthCheckCriticalTimeout: 90 * time.Second,
				})
			})

			Convey("Then a second call to config should return the same config", func() {
				newCfg, newErr := Get()
				So(newErr, ShouldBeNil)
				So(newCfg, ShouldResemble, cfg)
			})
		})
	})
}
