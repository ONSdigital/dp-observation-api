package service

import (
	"context"

	"github.com/ONSdigital/dp-graph/v2/graph"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	mongolib "github.com/ONSdigital/dp-mongodb"
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

// Service contains all the configs, server and clients to run the observation API
type Service struct {
	config      *config.Config
	server      *server.Server
	router      *mux.Router
	api         *api.API
	serviceList *initialise.ExternalServiceList
	healthCheck *healthcheck.HealthCheck
	mongodb     *mongo.Mongo
	graphDB     *graph.DB
}

//ObserverAPIStore is a wrapper which embeds Neo4j Mongo structs which between them satisfy the store.Storer interface.
type ObserverAPIStore struct {
	*mongo.Mongo
	*graph.DB
}

// Run the service with its dependencies
func Run(ctx context.Context, buildTime, gitCommit, version string, svcErrors chan error) (*Service, error) {
	log.Event(ctx, "running service", log.INFO)

	// Read config
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
	a := api.Setup(ctx, r, cfg, store)

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

	// Run the http server in a new go-routine
	go func() {
		if err := s.ListenAndServe(); err != nil {
			svcErrors <- errors.Wrap(err, "failure in http listen and serve")
		}
	}()

	return &Service{
		config:      cfg,
		router:      r,
		api:         a,
		healthCheck: &hc,
		server:      s,
		serviceList: &serviceList,
		mongodb:     mongodb,
		graphDB:     graphDB,
	}, nil
}

// Close gracefully shuts the service down in the required order, with timeout
func (svc *Service) Close(ctx context.Context) error {
	timeout := svc.config.GracefulShutdownTimeout
	log.Event(ctx, "commencing graceful shutdown", log.Data{"graceful_shutdown_timeout": timeout}, log.INFO)
	ctx, cancel := context.WithTimeout(ctx, timeout)

	// track shutdown gracefully closes app
	var gracefulShutdown bool

	go func() {
		defer cancel()
		var hasShutdownError bool

		// stop healthcheck, as it depends on everything else
		if svc.serviceList.HealthCheck {
			svc.healthCheck.Stop()
		}

		// stop any incoming requests before closing any outbound connections
		if err := svc.server.Shutdown(ctx); err != nil {
			log.Event(ctx, "failed to shutdown http server", log.Error(err), log.ERROR)
		}

		// close any API dependency
		if err := svc.api.Close(ctx); err != nil {
			log.Event(ctx, "error closing API", log.Error(err), log.ERROR)
		}

		// close mongodb
		if svc.serviceList.MongoDB {
			if err := mongolib.Close(ctx, svc.mongodb.Session); err != nil {
				log.Event(ctx, "failed to close mongo db session", log.ERROR, log.Error(err))
				hasShutdownError = true
			}
		}

		// close graph database
		if svc.serviceList.Graph {
			if err := svc.graphDB.Close(ctx); err != nil {
				log.Event(ctx, "failed to close graph db", log.ERROR, log.Error(err))
				hasShutdownError = true
			}
		}

		if !hasShutdownError {
			gracefulShutdown = true
		}
	}()

	// wait for shutdown success (via cancel) or failure (timeout)
	<-ctx.Done()

	if !gracefulShutdown {
		err := errors.New("failed to shutdown gracefully")
		log.Event(ctx, "failed to shutdown gracefully ", log.ERROR, log.Error(err))
		return err
	}

	log.Event(ctx, "graceful shutdown was successful", log.INFO)
	return nil
}

// registerCheckers adds the Checkers to the healthcheck client, for the provided dependencies
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
