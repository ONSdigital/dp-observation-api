package config

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfig(t *testing.T) {
	Convey("Given an environment with no environment variables set", t, func() {
		cfg, err := Get()

		Convey("When the config values are retrieved", func() {

			Convey("Then there should be no error returned", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then the values should be set to the expected defaults", func() {
				So(cfg, ShouldResemble, &Config{
					BindAddr:                   ":24500",
					GracefulShutdownTimeout:    5 * time.Second,
					HealthCheckInterval:        30 * time.Second,
					HealthCheckCriticalTimeout: 90 * time.Second,
					MongoConfig: MongoConfig{
						BindAddr:   "localhost:27017",
						Collection: "datasets",
						Database:   "datasets",
					},
				})
				// So(cfg.BindAddr, ShouldEqual, ":24500")
				// So(cfg.GracefulShutdownTimeout, ShouldEqual, 5*time.Second)
				// So(cfg.HealthCheckInterval, ShouldEqual, 30*time.Second)
				// So(cfg.HealthCheckCriticalTimeout, ShouldEqual, 90*time.Second)
				// So(cfg.EnablePrivateEnpoints, ShouldEqual, false)
				// So(cfg.MongoConfig, ShouldResemble, MongoConfig{
				// 	BindAddr:   "localhost:27017",
				// 	Collection: "datasets",
				// 	Database:   "datasets",
				// })
			})

			Convey("Then a second call to config should return the same config", func() {
				newCfg, newErr := Get()
				So(newErr, ShouldBeNil)
				So(newCfg, ShouldResemble, cfg)
			})
		})
	})
}
