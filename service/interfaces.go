package service

import (
	"context"
	"net/http"
	"time"

	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-observation-api/api"
	"github.com/ONSdigital/dp-observation-api/config"
)

//go:generate moq -out mock/initialiser.go -pkg mock . Initialiser
//go:generate moq -out mock/server.go -pkg mock . IServer
//go:generate moq -out mock/healthCheck.go -pkg mock . IHealthCheck
//go:generate moq -out mock/closer.go -pkg mock . Closer

// Initialiser defines the methods to initialise external services
type Initialiser interface {
	DoGetHTTPServer(bindAddr string, httpWriteTimeout time.Duration, router http.Handler) IServer
	DoGetGraphDB(ctx context.Context) (api.IGraph, Closer, error)
	DoGetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (IHealthCheck, error)
}

// IServer defines the required methods from the HTTP server
type IServer interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

// IHealthCheck defines the required methods from Healthcheck
type IHealthCheck interface {
	Handler(w http.ResponseWriter, req *http.Request)
	Start(ctx context.Context)
	Stop()
	AddCheck(name string, checker healthcheck.Checker) (err error)
}

type Closer interface {
	Close(ctx context.Context) error
}
