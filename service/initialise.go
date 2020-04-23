package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-graph/v2/graph"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-observation-api/config"
	"github.com/ONSdigital/dp-observation-api/mongo"
	"github.com/ONSdigital/go-ns/server"
	"github.com/ONSdigital/log.go/log"
)

// ExternalServiceList holds the initialiser and initialisation state of external services.
type ExternalServiceList struct {
	Graph       bool
	HealthCheck bool
	MongoDB     bool
	HTTPServer  bool
	Init        Initialiser
}

// NewServiceList creates a new service list with the provided initialiser
func NewServiceList(initialiser Initialiser) *ExternalServiceList {
	return &ExternalServiceList{
		Graph:       false,
		HealthCheck: false,
		MongoDB:     false,
		Init:        initialiser,
	}
}

// Init implements the Initialiser interface to initialise dependencies
type Init struct{}

// GetHTTPServer creates an http server and sets the Server flag to true
func (e *ExternalServiceList) GetHTTPServer(bindAddr string, router http.Handler) IServer {
	s := e.Init.DoGetHTTPServer(bindAddr, router)
	e.HTTPServer = true
	return s
}

// GetMongoDB creates a mongodb client and sets the MongoDB flag to true
func (e *ExternalServiceList) GetMongoDB(ctx context.Context, cfg *config.Config) (IMongo, error) {
	mongodb, err := e.Init.DoGetMongoDB(ctx, cfg)
	if err != nil {
		return nil, err
	}
	e.MongoDB = true
	return mongodb, nil
}

// GetGraphDB creates a graphDB client and sets the Graph flag to true
func (e *ExternalServiceList) GetGraphDB(ctx context.Context) (IGraph, error) {
	graphDB, err := e.Init.DoGetGraphDB(ctx)
	if err != nil {
		return nil, err
	}
	e.Graph = true
	return graphDB, nil
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
func (e *Init) DoGetHTTPServer(bindAddr string, router http.Handler) IServer {
	s := server.New(bindAddr, router)
	s.HandleOSSignals = false
	return s
}

// DoGetMongoDB returns a mongodb client and dataset mongo object
func (e *Init) DoGetMongoDB(ctx context.Context, cfg *config.Config) (IMongo, error) {
	mongodb := &mongo.Mongo{
		Collection: cfg.MongoConfig.Collection,
		Database:   cfg.MongoConfig.Database,
		URI:        cfg.MongoConfig.BindAddr,
	}

	// Initialise mongoDB sessions and client
	_, err := mongodb.Init()
	if err != nil {
		log.Event(ctx, "failed to initialise mongo", log.ERROR, log.Error(err))
		return nil, err
	}
	log.Event(ctx, "listening to mongo db session", log.INFO, log.Data{
		"bind_address": cfg.BindAddr,
	})

	return mongodb, nil
}

// DoGetGraphDB returns a graphDB
func (e *Init) DoGetGraphDB(ctx context.Context) (IGraph, error) {
	return graph.New(ctx, graph.Subsets{Observation: true, Instance: true})
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
