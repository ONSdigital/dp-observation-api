package service

import (
	"context"
	"net/http"

	"github.com/globalsign/mgo"

	"github.com/ONSdigital/dp-graph/v2/graph/driver"
	"github.com/ONSdigital/dp-graph/v2/observation"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-observation-api/config"
	"github.com/ONSdigital/dp-observation-api/models"
)

//go:generate moq -out mock/initialiser.go -pkg mock . Initialiser
//go:generate moq -out mock/server.go -pkg mock . IServer
//go:generate moq -out mock/healthCheck.go -pkg mock . IHealthCheck
//go:generate moq -out mock/mongo.go -pkg mock . IMongo
//go:generate moq -out mock/graph.go -pkg mock . IGraph

// Initialiser defines the methods to initialise external services
type Initialiser interface {
	DoGetHTTPServer(bindAddr string, router http.Handler) IServer
	DoGetMongoDB(ctx context.Context, cfg *config.Config) (IMongo, error)
	DoGetGraphDB(ctx context.Context) (IGraph, error)
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

// IMongo defines the required methods from MongoDB
type IMongo interface {
	GetDataset(ID string) (*models.DatasetUpdate, error)
	CheckEditionExists(ID, editionID, state string) error
	GetVersion(datasetID, editionID, version, state string) (*models.Version, error)
	Session() *mgo.Session
	Close(ctx context.Context) error
	Checker(ctx context.Context, state *healthcheck.CheckState) error
}

// IGraph defines the required methods from GraphDB
type IGraph interface {
	driver.Driver
	StreamCSVRows(ctx context.Context, instanceID, filterID string, filters *observation.DimensionFilters, limit *int) (observation.StreamRowReader, error)
}
