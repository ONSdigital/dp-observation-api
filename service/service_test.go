package service_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-observation-api/config"
	"github.com/ONSdigital/dp-observation-api/service"
	"github.com/ONSdigital/dp-observation-api/service/mock"
	"github.com/globalsign/mgo"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	ctx           = context.Background()
	testBuildTime = "BuildTime"
	testGitCommit = "GitCommit"
	testVersion   = "Version"
)

var (
	errMongo       = errors.New("mongoDB error")
	errGraph       = errors.New("graphDB error")
	errHealthcheck = errors.New("healthCheck error")
)

var funcDoGetMongoDbErr = func(ctx context.Context, cfg *config.Config) (service.IMongo, error) {
	return nil, errMongo
}

var funcDoGetGraphDbErr = func(ctx context.Context) (service.IGraph, error) {
	return nil, errGraph
}

var funcDoGetHealthcheckErr = func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.IHealthCheck, error) {
	return nil, errHealthcheck
}

var funcDoGetHTTPServerNil = func(bindAddr string, router http.Handler) service.IServer {
	return nil
}

func createMongoMock() {

}

func TestRun(t *testing.T) {

	Convey("Having a set of mocked dependencies", t, func() {

		mongoDbMock := &mock.IMongoMock{
			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
		}

		graphDbMock := &mock.IGraphMock{
			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
		}

		hcMock := &mock.IHealthCheckMock{
			AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
			StartFunc:    func(ctx context.Context) {},
		}

		serverMock := &mock.IServerMock{
			ListenAndServeFunc: func() error { return nil },
		}

		funcDoGetMongoDbOk := func(ctx context.Context, cfg *config.Config) (service.IMongo, error) {
			return mongoDbMock, nil
		}

		funcDoGetGraphDbOk := func(ctx context.Context) (service.IGraph, error) {
			return graphDbMock, nil
		}

		funcDoGetHealthcheckOk := func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.IHealthCheck, error) {
			return hcMock, nil
		}

		funcDoGetHTTPServer := func(bindAddr string, router http.Handler) service.IServer {
			return serverMock
		}

		Convey("Given that initialising mongoDB returns an error", func() {
			initMock := &mock.InitialiserMock{
				DoGetHTTPServerFunc: funcDoGetHTTPServerNil,
				DoGetMongoDBFunc:    funcDoGetMongoDbErr,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			_, err := service.Run(ctx, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails with the same error and the flag is not set", func() {
				So(err, ShouldResemble, errMongo)
				So(svcList.HTTPServer, ShouldBeTrue)
				So(svcList.MongoDB, ShouldBeFalse)
				So(svcList.Graph, ShouldBeFalse)
				So(svcList.HealthCheck, ShouldBeFalse)
			})
		})

		Convey("Given that initialising graphDB returns an error", func() {
			initMock := &mock.InitialiserMock{
				DoGetHTTPServerFunc: funcDoGetHTTPServerNil,
				DoGetMongoDBFunc:    funcDoGetMongoDbOk,
				DoGetGraphDBFunc:    funcDoGetGraphDbErr,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			_, err := service.Run(ctx, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails with the same error and the flag is not set", func() {
				So(err, ShouldResemble, errGraph)
				So(svcList.MongoDB, ShouldBeTrue)
				So(svcList.Graph, ShouldBeFalse)
				So(svcList.HealthCheck, ShouldBeFalse)
			})
		})

		Convey("Given that initialising healthcheck returns an error", func() {
			initMock := &mock.InitialiserMock{
				DoGetHTTPServerFunc:  funcDoGetHTTPServerNil,
				DoGetMongoDBFunc:     funcDoGetMongoDbOk,
				DoGetGraphDBFunc:     funcDoGetGraphDbOk,
				DoGetHealthCheckFunc: funcDoGetHealthcheckErr,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			_, err := service.Run(ctx, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails with the same error and the flag is not set", func() {
				So(err, ShouldResemble, errHealthcheck)
				So(svcList.MongoDB, ShouldBeTrue)
				So(svcList.Graph, ShouldBeTrue)
				So(svcList.HealthCheck, ShouldBeFalse)
			})
		})

		Convey("Given that all dependencies are successfully initialised", func() {

			initMock := &mock.InitialiserMock{
				DoGetHTTPServerFunc:  funcDoGetHTTPServer,
				DoGetMongoDBFunc:     funcDoGetMongoDbOk,
				DoGetGraphDBFunc:     funcDoGetGraphDbOk,
				DoGetHealthCheckFunc: funcDoGetHealthcheckOk,
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			_, err := service.Run(ctx, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run succeeds and all the flags are set", func() {
				So(err, ShouldBeNil)
				So(svcList.MongoDB, ShouldBeTrue)
				So(svcList.Graph, ShouldBeTrue)
				So(svcList.HealthCheck, ShouldBeTrue)
			})

			Convey("The checkers are registered and the healthcheck and http server started", func() {
				So(len(hcMock.AddCheckCalls()), ShouldEqual, 3)
				So(hcMock.AddCheckCalls()[0].Name, ShouldResemble, "Graph DB")
				So(hcMock.AddCheckCalls()[1].Name, ShouldResemble, "Mongo DB")
				So(hcMock.AddCheckCalls()[2].Name, ShouldResemble, "Dataset API")
				So(len(initMock.DoGetHTTPServerCalls()), ShouldEqual, 1)
				So(initMock.DoGetHTTPServerCalls()[0].BindAddr, ShouldEqual, ":24500")
				So(len(hcMock.StartCalls()), ShouldEqual, 1)
				So(len(serverMock.ListenAndServeCalls()), ShouldEqual, 1)
			})
		})

		Convey("Given that Checkers cannot be registered", func() {

			errAddheckFail := errors.New("Error(s) registering checkers for healthcheck")
			hcMockAddFail := &mock.IHealthCheckMock{
				AddCheckFunc: func(name string, checker healthcheck.Checker) error { return errAddheckFail },
				StartFunc:    func(ctx context.Context) {},
			}

			initMock := &mock.InitialiserMock{
				DoGetHTTPServerFunc: funcDoGetHTTPServerNil,
				DoGetMongoDBFunc:    funcDoGetMongoDbOk,
				DoGetGraphDBFunc:    funcDoGetGraphDbOk,
				DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.IHealthCheck, error) {
					return hcMockAddFail, nil
				},
			}
			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			_, err := service.Run(ctx, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)

			Convey("Then service Run fails, but all checks try to register", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldResemble, fmt.Sprintf("unable to register checkers: %s", errAddheckFail.Error()))
				So(svcList.MongoDB, ShouldBeTrue)
				So(svcList.Graph, ShouldBeTrue)
				So(svcList.HealthCheck, ShouldBeTrue)
				So(len(hcMockAddFail.AddCheckCalls()), ShouldEqual, 3)
				So(hcMockAddFail.AddCheckCalls()[0].Name, ShouldResemble, "Graph DB")
				So(hcMockAddFail.AddCheckCalls()[1].Name, ShouldResemble, "Mongo DB")
				So(hcMockAddFail.AddCheckCalls()[2].Name, ShouldResemble, "Dataset API")
			})
		})
	})
}

func TestClose(t *testing.T) {

	Convey("Having a correctly initialised service", t, func() {

		hcStopped := false
		serverStopped := false

		// mongoDB Close will fail if healthcheck and http server
		mongoDbMock := &mock.IMongoMock{
			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
			SessionFunc: func() *mgo.Session { return nil },
			CloseFunc: func(ctx context.Context) error {
				if !hcStopped || !serverStopped {
					return errors.New("MongoDB closed before stopping healthcheck or HTTP server")
				}
				return nil
			},
		}

		// graphDB Close will fail if healthcheck and http server
		graphDbMock := &mock.IGraphMock{
			CheckerFunc: func(ctx context.Context, state *healthcheck.CheckState) error { return nil },
			CloseFunc: func(ctx context.Context) error {
				if !hcStopped || !serverStopped {
					return errors.New("GraphDB closed before stopping healthcheck or HTTP server")
				}
				return nil
			},
		}

		// healthcheck Stop does not depend on any other service being closed/stopped
		hcMock := &mock.IHealthCheckMock{
			AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
			StartFunc:    func(ctx context.Context) {},
			StopFunc:     func() { hcStopped = true },
		}

		// server Shutdown will fail if healthcheck is not stopped
		serverMock := &mock.IServerMock{
			ListenAndServeFunc: func() error { return nil },
			ShutdownFunc: func(ctx context.Context) error {
				if !hcStopped {
					return errors.New("Server stopped before healthcheck")
				}
				serverStopped = true
				return nil
			},
		}

		Convey("Closing the service results in all the dependencies being closed in the expected order", func() {

			initMock := &mock.InitialiserMock{
				DoGetHTTPServerFunc: func(bindAddr string, router http.Handler) service.IServer { return serverMock },
				DoGetMongoDBFunc:    func(ctx context.Context, cfg *config.Config) (service.IMongo, error) { return mongoDbMock, nil },
				DoGetGraphDBFunc:    func(ctx context.Context) (service.IGraph, error) { return graphDbMock, nil },
				DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.IHealthCheck, error) {
					return hcMock, nil
				},
			}

			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			svc, err := service.Run(ctx, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)
			So(err, ShouldBeNil)

			err = svc.Close(context.Background())
			So(err, ShouldBeNil)
			So(len(hcMock.StopCalls()), ShouldEqual, 1)
			So(len(serverMock.ShutdownCalls()), ShouldEqual, 1)
			So(len(graphDbMock.CloseCalls()), ShouldEqual, 1)
			So(len(mongoDbMock.CloseCalls()), ShouldEqual, 1)
		})

		Convey("If services fail to stop, the Close operation tries to close all dependencies and returns an error", func() {

			failingserverMock := &mock.IServerMock{
				ListenAndServeFunc: func() error { return nil },
				ShutdownFunc: func(ctx context.Context) error {
					return errors.New("Failed to stop http server")
				},
			}

			initMock := &mock.InitialiserMock{
				DoGetHTTPServerFunc: func(bindAddr string, router http.Handler) service.IServer { return failingserverMock },
				DoGetMongoDBFunc:    func(ctx context.Context, cfg *config.Config) (service.IMongo, error) { return mongoDbMock, nil },
				DoGetGraphDBFunc:    func(ctx context.Context) (service.IGraph, error) { return graphDbMock, nil },
				DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.IHealthCheck, error) {
					return hcMock, nil
				},
			}

			svcErrors := make(chan error, 1)
			svcList := service.NewServiceList(initMock)
			svc, err := service.Run(ctx, svcList, testBuildTime, testGitCommit, testVersion, svcErrors)
			So(err, ShouldBeNil)

			err = svc.Close(context.Background())
			So(err, ShouldNotBeNil)
			So(len(hcMock.StopCalls()), ShouldEqual, 1)
			So(len(failingserverMock.ShutdownCalls()), ShouldEqual, 1)
			So(len(graphDbMock.CloseCalls()), ShouldEqual, 1)
			So(len(mongoDbMock.CloseCalls()), ShouldEqual, 1)
		})
	})
}
