package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-authorisation/auth"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	rchttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/dp-observation-api/api"
	"github.com/ONSdigital/dp-observation-api/config"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// Service contains all the configs, server and clients to run the observation API
type Service struct {
	config             *config.Config
	server             IServer
	router             *mux.Router
	api                *api.API
	serviceList        *ExternalServiceList
	healthCheck        IHealthCheck
	graphDB            api.IGraph
	graphErrorConsumer Closer
	cantabularClient   CantabularClient
}

// Run the service with its dependencies
func Run(ctx context.Context, cfg *config.Config, serviceList *ExternalServiceList, buildTime, gitCommit, version string, svcErrors chan error) (*Service, error) {
	log.Event(ctx, "running service", log.INFO)

	log.Event(ctx, "got service configuration", log.Data{"config": cfg}, log.INFO)

	// Get HTTP Server
	r := mux.NewRouter()
	s := serviceList.GetHTTPServer(cfg.BindAddr, cfg.HttpWriteTimeout, r)

	// Get graphDB connection for observation store
	graphDB, graphErrorConsumer, err := serviceList.GetGraphDB(ctx)
	if err != nil {
		log.Event(ctx, "failed to initialise graph driver", log.FATAL, log.Error(err))
		return nil, err
	}

	// Get zebedee client
	zebedeeCli := zebedee.New(cfg.ZebedeeURL)

	// Get dataset API client
	datasetAPICli := dataset.NewAPIClient(cfg.DatasetAPIURL)

	cantabularClient := serviceList.GetCantabularClient(ctx, cfg)

	// Get permissions for private endpoints
	permissions := getAuthorisationHandler(ctx, *cfg)

	// Get HealthCheck
	hc, err := serviceList.GetHealthCheck(cfg, buildTime, gitCommit, version)
	if err != nil {
		log.Event(ctx, "could not instantiate healthcheck", log.FATAL, log.Error(err))
		return nil, err
	}
	if err := registerCheckers(ctx, cfg, hc, graphDB, zebedeeCli, datasetAPICli, cantabularClient, cfg.EnablePrivateEndpoints); err != nil {
		return nil, errors.Wrap(err, "unable to register checkers")
	}

	r.StrictSlash(true).Path("/health").HandlerFunc(hc.Handler)
	hc.Start(ctx)

	// Setup the API
	a := api.Setup(ctx, r, cfg, graphDB, datasetAPICli, cantabularClient, permissions)

	// Run the http server in a new go-routine
	go func() {
		if err := s.ListenAndServe(); err != nil {
			svcErrors <- errors.Wrap(err, "failure in http listen and serve")
		}
	}()

	return &Service{
		config:             cfg,
		router:             r,
		api:                a,
		healthCheck:        hc,
		server:             s,
		serviceList:        serviceList,
		graphDB:            graphDB,
		graphErrorConsumer: graphErrorConsumer,
		cantabularClient:   cantabularClient,
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
		if svc.serviceList.HTTPServer {
			if err := svc.server.Shutdown(ctx); err != nil {
				log.Event(ctx, "failed to shutdown http server", log.Error(err), log.ERROR)
			}
		}

		// close API
		if err := svc.api.Close(ctx); err != nil {
			log.Event(ctx, "error closing API", log.Error(err), log.ERROR)
		}

		// close graph database
		if svc.serviceList.Graph {
			if err := svc.graphDB.Close(ctx); err != nil {
				log.Event(ctx, "failed to close graph db", log.ERROR, log.Error(err))
				hasShutdownError = true
			}

			if err := svc.graphErrorConsumer.Close(ctx); err != nil {
				log.Event(ctx, "failed to close graph db error consumer", log.ERROR, log.Error(err))
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

// getAuthorisationHandler retrieves auth handler to authorise request
func getAuthorisationHandler(ctx context.Context, cfg config.Config) api.IAuthHandler {
	if !cfg.EnablePrivateEndpoints {
		log.Event(ctx, "feature flag to not enable private endpoints, nop auth impl", log.INFO, log.Data{"feature": "ENABLE_PRIVATE_ENDPOINTS"})
		return &auth.NopHandler{}
	}

	log.Event(ctx, "feature flag enabled", log.INFO, log.Data{"feature": "ENABLE_PERMISSIONS_AUTH"})
	auth.LoggerNamespace("dp-observation-api-auth")

	// for checking caller permissions when we only have a user/service token
	return auth.NewHandler(
		auth.NewDatasetPermissionsRequestBuilder(cfg.ZebedeeURL, "dataset_id", mux.Vars),
		auth.NewPermissionsClient(rchttp.NewClient()),
		auth.DefaultPermissionsVerifier(),
	)
}

// registerCheckers adds the Checkers to the healthcheck client, for the provided dependencies
func registerCheckers(ctx context.Context, cfg *config.Config,
	hc IHealthCheck,
	graphDB api.IGraph,
	zebedeeCli *zebedee.Client,
	datasetAPICli api.IDatasetClient,
	cantabularClient CantabularClient,
	enablePrivateEndpoints bool) (err error) {

	hasErrors := false

	if enablePrivateEndpoints {
		if err = hc.AddCheck("Zebedee", zebedeeCli.Checker); err != nil {
			hasErrors = true
			log.Event(ctx, "error adding check for zebedee", log.ERROR, log.Error(err))
		}
	}

	if err = hc.AddCheck("Graph DB", graphDB.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "error adding check for graph db", log.ERROR, log.Error(err))
	}

	if err = hc.AddCheck("Dataset API", datasetAPICli.Checker); err != nil {
		hasErrors = true
		log.Event(ctx, "error adding check for dataset api", log.ERROR, log.Error(err))
	}

	cantabularChecker := cantabularClient.Checker
	if !cfg.CantabularHealthcheckEnabled {
		cantabularChecker = func(ctx context.Context, state *healthcheck.CheckState) error {
			state.Update(healthcheck.StatusOK, "Cantabular healthcheck placeholder", http.StatusOK)
			return nil
		}
	}

	if err = hc.AddCheck("Dataset API", cantabularChecker); err != nil {
		hasErrors = true
		log.Event(ctx, "error adding check for dataset api", log.ERROR, log.Error(err))
	}

	if hasErrors {
		return errors.New("Error(s) registering checkers for healthcheck")
	}
	return nil
}
