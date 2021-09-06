// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package mock

import (
	"context"
	"github.com/ONSdigital/dp-observation-api/api"
	"github.com/ONSdigital/dp-observation-api/config"
	"github.com/ONSdigital/dp-observation-api/service"
	"net/http"
	"sync"
	"time"
)

// Ensure, that InitialiserMock does implement service.Initialiser.
// If this is not the case, regenerate this file with moq.
var _ service.Initialiser = &InitialiserMock{}

// InitialiserMock is a mock implementation of service.Initialiser.
//
// 	func TestSomethingThatUsesInitialiser(t *testing.T) {
//
// 		// make and configure a mocked service.Initialiser
// 		mockedInitialiser := &InitialiserMock{
// 			DoGetGraphDBFunc: func(ctx context.Context) (api.IGraph, service.Closer, error) {
// 				panic("mock out the DoGetGraphDB method")
// 			},
// 			DoGetHTTPServerFunc: func(bindAddr string, httpWriteTimeout time.Duration, router http.Handler) service.IServer {
// 				panic("mock out the DoGetHTTPServer method")
// 			},
// 			DoGetHealthCheckFunc: func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.IHealthCheck, error) {
// 				panic("mock out the DoGetHealthCheck method")
// 			},
// 		}
//
// 		// use mockedInitialiser in code that requires service.Initialiser
// 		// and then make assertions.
//
// 	}
type InitialiserMock struct {
	// DoGetGraphDBFunc mocks the DoGetGraphDB method.
	DoGetGraphDBFunc func(ctx context.Context) (api.IGraph, service.Closer, error)

	// DoGetHTTPServerFunc mocks the DoGetHTTPServer method.
	DoGetHTTPServerFunc func(bindAddr string, httpWriteTimeout time.Duration, router http.Handler) service.IServer

	// DoGetHealthCheckFunc mocks the DoGetHealthCheck method.
	DoGetHealthCheckFunc func(cfg *config.Config, buildTime string, gitCommit string, version string) (service.IHealthCheck, error)

	// calls tracks calls to the methods.
	calls struct {
		// DoGetGraphDB holds details about calls to the DoGetGraphDB method.
		DoGetGraphDB []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
		// DoGetHTTPServer holds details about calls to the DoGetHTTPServer method.
		DoGetHTTPServer []struct {
			// BindAddr is the bindAddr argument value.
			BindAddr string
			// HttpWriteTimeout is the httpWriteTimeout argument value.
			HttpWriteTimeout time.Duration
			// Router is the router argument value.
			Router http.Handler
		}
		// DoGetHealthCheck holds details about calls to the DoGetHealthCheck method.
		DoGetHealthCheck []struct {
			// Cfg is the cfg argument value.
			Cfg *config.Config
			// BuildTime is the buildTime argument value.
			BuildTime string
			// GitCommit is the gitCommit argument value.
			GitCommit string
			// Version is the version argument value.
			Version string
		}
	}
	lockDoGetGraphDB     sync.RWMutex
	lockDoGetHTTPServer  sync.RWMutex
	lockDoGetHealthCheck sync.RWMutex
}

// DoGetGraphDB calls DoGetGraphDBFunc.
func (mock *InitialiserMock) DoGetGraphDB(ctx context.Context) (api.IGraph, service.Closer, error) {
	if mock.DoGetGraphDBFunc == nil {
		panic("InitialiserMock.DoGetGraphDBFunc: method is nil but Initialiser.DoGetGraphDB was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.lockDoGetGraphDB.Lock()
	mock.calls.DoGetGraphDB = append(mock.calls.DoGetGraphDB, callInfo)
	mock.lockDoGetGraphDB.Unlock()
	return mock.DoGetGraphDBFunc(ctx)
}

// DoGetGraphDBCalls gets all the calls that were made to DoGetGraphDB.
// Check the length with:
//     len(mockedInitialiser.DoGetGraphDBCalls())
func (mock *InitialiserMock) DoGetGraphDBCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	mock.lockDoGetGraphDB.RLock()
	calls = mock.calls.DoGetGraphDB
	mock.lockDoGetGraphDB.RUnlock()
	return calls
}

// DoGetHTTPServer calls DoGetHTTPServerFunc.
func (mock *InitialiserMock) DoGetHTTPServer(bindAddr string, httpWriteTimeout time.Duration, router http.Handler) service.IServer {
	if mock.DoGetHTTPServerFunc == nil {
		panic("InitialiserMock.DoGetHTTPServerFunc: method is nil but Initialiser.DoGetHTTPServer was just called")
	}
	callInfo := struct {
		BindAddr         string
		HttpWriteTimeout time.Duration
		Router           http.Handler
	}{
		BindAddr:         bindAddr,
		HttpWriteTimeout: httpWriteTimeout,
		Router:           router,
	}
	mock.lockDoGetHTTPServer.Lock()
	mock.calls.DoGetHTTPServer = append(mock.calls.DoGetHTTPServer, callInfo)
	mock.lockDoGetHTTPServer.Unlock()
	return mock.DoGetHTTPServerFunc(bindAddr, httpWriteTimeout, router)
}

// DoGetHTTPServerCalls gets all the calls that were made to DoGetHTTPServer.
// Check the length with:
//     len(mockedInitialiser.DoGetHTTPServerCalls())
func (mock *InitialiserMock) DoGetHTTPServerCalls() []struct {
	BindAddr         string
	HttpWriteTimeout time.Duration
	Router           http.Handler
} {
	var calls []struct {
		BindAddr         string
		HttpWriteTimeout time.Duration
		Router           http.Handler
	}
	mock.lockDoGetHTTPServer.RLock()
	calls = mock.calls.DoGetHTTPServer
	mock.lockDoGetHTTPServer.RUnlock()
	return calls
}

// DoGetHealthCheck calls DoGetHealthCheckFunc.
func (mock *InitialiserMock) DoGetHealthCheck(cfg *config.Config, buildTime string, gitCommit string, version string) (service.IHealthCheck, error) {
	if mock.DoGetHealthCheckFunc == nil {
		panic("InitialiserMock.DoGetHealthCheckFunc: method is nil but Initialiser.DoGetHealthCheck was just called")
	}
	callInfo := struct {
		Cfg       *config.Config
		BuildTime string
		GitCommit string
		Version   string
	}{
		Cfg:       cfg,
		BuildTime: buildTime,
		GitCommit: gitCommit,
		Version:   version,
	}
	mock.lockDoGetHealthCheck.Lock()
	mock.calls.DoGetHealthCheck = append(mock.calls.DoGetHealthCheck, callInfo)
	mock.lockDoGetHealthCheck.Unlock()
	return mock.DoGetHealthCheckFunc(cfg, buildTime, gitCommit, version)
}

// DoGetHealthCheckCalls gets all the calls that were made to DoGetHealthCheck.
// Check the length with:
//     len(mockedInitialiser.DoGetHealthCheckCalls())
func (mock *InitialiserMock) DoGetHealthCheckCalls() []struct {
	Cfg       *config.Config
	BuildTime string
	GitCommit string
	Version   string
} {
	var calls []struct {
		Cfg       *config.Config
		BuildTime string
		GitCommit string
		Version   string
	}
	mock.lockDoGetHealthCheck.RLock()
	calls = mock.calls.DoGetHealthCheck
	mock.lockDoGetHealthCheck.RUnlock()
	return calls
}
