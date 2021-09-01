package service

import (
	"context"
	"net/http"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/cantabular"
	"github.com/ONSdigital/dp-graph/v2/graph"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dpHTTP "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/dp-observation-api/api"
	"github.com/ONSdigital/dp-observation-api/config"
)

// ExternalServiceList holds the initialiser and initialisation state of external services.
type ExternalServiceList struct {
	Graph       bool
	HealthCheck bool
	HTTPServer  bool
	Init        Initialiser
}

// NewServiceList creates a new service list with the provided initialiser
func NewServiceList(initialiser Initialiser) *ExternalServiceList {
	return &ExternalServiceList{
		Graph:       false,
		HealthCheck: false,
		Init:        initialiser,
	}
}

// Init implements the Initialiser interface to initialise dependencies
type Init struct{}

// GetHTTPServer creates an http server and sets the Server flag to true
func (e *ExternalServiceList) GetHTTPServer(bindAddr string, httpWriteTimeout time.Duration, router http.Handler) IServer {
	s := e.Init.DoGetHTTPServer(bindAddr, httpWriteTimeout, router)
	e.HTTPServer = true
	return s
}

// GetGraphDB creates a graphDB client and sets the Graph flag to true
func (e *ExternalServiceList) GetGraphDB(ctx context.Context) (api.IGraph, Closer, error) {
	graphDB, graphDBErrorConsumer, err := e.Init.DoGetGraphDB(ctx)
	if err != nil {
		return nil, nil, err
	}
	e.Graph = true
	return graphDB, graphDBErrorConsumer, nil
}

// GetHealthCheck creates a healthcheck with versionInfo and sets teh HealthCheck flag to true
func (e *ExternalServiceList) GetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (IHealthCheck, error) {
	hc, err := e.Init.DoGetHealthCheck(cfg, buildTime, gitCommit, version)
	if err != nil {
		return nil, err
	}
	e.HealthCheck = true
	return hc, nil
}

// DoGetHTTPServer creates an HTTP Server with the provided bind address and router
func (e *Init) DoGetHTTPServer(bindAddr string, httpWriteTimeout time.Duration, router http.Handler) IServer {
	s := dpHTTP.NewServer(bindAddr, router)
	s.WriteTimeout = httpWriteTimeout
	s.HandleOSSignals = false
	return s
}

// DoGetGraphDB returns a graphDB
func (e *Init) DoGetGraphDB(ctx context.Context) (api.IGraph, Closer, error) {
	graphDB, err := graph.New(ctx, graph.Subsets{Observation: true, Instance: true})
	if err != nil {
		return nil, nil, err
	}

	graphErrorConsumer := graph.NewLoggingErrorConsumer(ctx, graphDB.ErrorChan())

	return graphDB, graphErrorConsumer, nil
}

// DoGetHealthCheck creates a healthcheck with versionInfo
func (e *Init) DoGetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (IHealthCheck, error) {
	versionInfo, err := healthcheck.NewVersionInfo(buildTime, gitCommit, version)
	if err != nil {
		return nil, err
	}
	hc := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)
	return &hc, nil
}

func (e *ExternalServiceList) GetCantabularClient(ctx context.Context, cfg *config.Config) CantabularClient {
	return cantabular.NewClient(
		cantabular.Config{
			Host:           cfg.CantabularURL,
			ExtApiHost:     cfg.CantabularExtURL,
			GraphQLTimeout: cfg.DefaultRequestTimeout,
		},
		dpHTTP.NewClient(),
		nil,
	)
}
