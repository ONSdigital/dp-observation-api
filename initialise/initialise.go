package initialise

import (
	"context"

	"github.com/ONSdigital/dp-graph/v2/graph"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	mongoHealth "github.com/ONSdigital/dp-mongodb/health"
	"github.com/ONSdigital/dp-observation-api/config"
	"github.com/ONSdigital/dp-observation-api/mongo"
	"github.com/ONSdigital/log.go/log"
)

// ExternalServiceList represents a list of services
type ExternalServiceList struct {
	Graph       bool
	HealthCheck bool
	MongoDB     bool
}

// GetMongoDB returns a mongodb client and dataset mongo object
func (e *ExternalServiceList) GetMongoDB(ctx context.Context, cfg *config.Config) (*mongoHealth.Client, *mongo.Mongo, error) {
	mongodb := &mongo.Mongo{
		Collection: cfg.MongoConfig.Collection,
		Database:   cfg.MongoConfig.Database,
		URI:        cfg.MongoConfig.BindAddr,
	}

	session, err := mongodb.Init()
	if err != nil {
		log.Event(ctx, "failed to initialise mongo", log.ERROR, log.Error(err))
		return nil, nil, err
	}
	mongodb.Session = session
	log.Event(ctx, "listening to mongo db session", log.INFO, log.Data{
		"bind_address": cfg.BindAddr,
	})

	client := mongoHealth.NewClient(session)

	e.MongoDB = true

	return client, mongodb, nil
}

// GetGraphDB returns a graphDB
func (e *ExternalServiceList) GetGraphDB(ctx context.Context) (*graph.DB, error) {

	graphDB, err := graph.New(ctx, graph.Subsets{Observation: true, Instance: true})
	if err != nil {
		return nil, err
	}

	e.Graph = true

	return graphDB, nil
}

// GetHealthCheck creates a healthcheck with versionInfo
func (e *ExternalServiceList) GetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (healthcheck.HealthCheck, error) {

	// Create healthcheck object with versionInfo
	versionInfo, err := healthcheck.NewVersionInfo(buildTime, gitCommit, version)
	if err != nil {
		return healthcheck.HealthCheck{}, err
	}
	hc := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)

	e.HealthCheck = true

	return hc, nil
}
