package service

import (
	"context"

	"github.com/ONSdigital/dp-graph/v2/graph"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	mongoHealth "github.com/ONSdigital/dp-mongodb/health"
	"github.com/ONSdigital/dp-observation-api/api"
	"github.com/ONSdigital/dp-observation-api/config"
	"github.com/ONSdigital/dp-observation-api/initialise"
	"github.com/ONSdigital/dp-observation-api/mongo"
	"github.com/ONSdigital/dp-observation-api/store"
	"github.com/ONSdigital/go-ns/server"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// Service contains all the configs and clients to run the observation API
type Service struct {
	Config      *config.Config
	server      *server.Server
	Router      *mux.Router
	API         *api.API
	HealthCheck *healthcheck.HealthCheck
}

//ObserverAPIStore is a wrapper which embeds Neo4j Mongo structs which between them satisfy the store.Storer interface.
type ObserverAPIStore struct {
	*mongo.Mongo
	*graph.DB
}

// Run the service with its dependencies
func Run(buildTime, gitCommit, version string, svcErrors chan error) (*Service, error) {
	ctx := context.Background()
	log.Event(ctx, "running service", log.INFO)

	cfg, err := config.Get()
	if err != nil {
		return nil, errors.Wrap(err, "unable to retrieve service configuration")
	}
	log.Event(ctx, "got service configuration", log.Data{"config": cfg}, log.INFO)

	// External services and their initialization state
	var serviceList initialise.ExternalServiceList

	r := mux.NewRouter()

	s := server.New(cfg.BindAddr, r)
	s.HandleOSSignals = false

	// Get mongoDB connection for observation store
	mongoClient, mongodb, err := serviceList.GetMongoDB(ctx, cfg)
	if err != nil {
		log.Event(ctx, "could not obtain mongo session", log.ERROR, log.Error(err))
		return nil, err
	}

	// Get graphDB connection for observation store
	graphDB, err := serviceList.GetGraphDB(ctx)
	if err != nil {
		log.Event(ctx, "failed to initialise graph driver", log.FATAL, log.Error(err))
		return nil, err
	}

	store := store.DataStore{Backend: ObserverAPIStore{mongodb, graphDB}}

	// Setup the API
	a := api.Setup(ctx, r, store)

	// Get HealthCheck
	hc, err := serviceList.GetHealthCheck(cfg, buildTime, gitCommit, version)
	if err != nil {
		log.Event(ctx, "could not instantiate healthcheck", log.FATAL, log.Error(err))
		return nil, err
	}
	if err := registerCheckers(ctx, &hc, graphDB, mongoClient); err != nil {
		return nil, errors.Wrap(err, "unable to register checkers")
	}

	r.StrictSlash(true).Path("/health").HandlerFunc(hc.Handler)
	hc.Start(ctx)

	go func() {
		if err := s.ListenAndServe(); err != nil {
			svcErrors <- errors.Wrap(err, "failure in http listen and serve")
		}
	}()

	return &Service{
		Config:      cfg,
		Router:      r,
		API:         a,
		HealthCheck: &hc,
		server:      s,
	}, nil
}

// Gracefully shutdown the service
func (svc *Service) Close(ctx context.Context) {
	timeout := svc.Config.GracefulShutdownTimeout
	log.Event(ctx, "commencing graceful shutdown", log.Data{"graceful_shutdown_timeout": timeout}, log.INFO)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// stop any incoming requests before closing any outbound connections
	if err := svc.server.Shutdown(ctx); err != nil {
		log.Event(ctx, "failed to shutdown http server", log.Error(err), log.ERROR)
	}

	if err := svc.API.Close(ctx); err != nil {
		log.Event(ctx, "error closing API", log.Error(err), log.ERROR)
	}

	log.Event(ctx, "graceful shutdown complete", log.INFO)
}

func registerCheckers(ctx context.Context,
	hc *healthcheck.HealthCheck,
	graphDB *graph.DB,
	mongoClient *mongoHealth.Client) (err error) {

	hasErrors := false

	if err = hc.AddCheck("Graph DB", graphDB.Driver.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "error adding check for graph db", log.ERROR, log.Error(err))
	}

	mongoHealth := mongoHealth.CheckMongoClient{
		Client:      *mongoClient,
		Healthcheck: mongoClient.Healthcheck,
	}
	if err = hc.AddCheck("Mongo DB", mongoHealth.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "error adding check for mongo db", log.ERROR, log.Error(err))
	}

	if hasErrors {
		return errors.New("Error(s) registering checkers for healthcheck")
	}
	return nil
}
