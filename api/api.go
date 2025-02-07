package api

import (
	"context"
	"net/http"
	"net/url"

	"github.com/ONSdigital/dp-authorisation/auth"
	"github.com/ONSdigital/dp-net/request"
	"github.com/ONSdigital/dp-observation-api/config"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
)

// API provides a struct to wrap the api around
type API struct {
	cfg                *config.Config
	Router             *mux.Router
	graphDB            IGraph
	datasetClient      IDatasetClient
	cantabularClient   CantabularClient
	permissions        IAuthHandler
	enableURLRewriting bool
	observationAPIURL  *url.URL
}

// Setup creates the API struct and its endpoints with corresponding handlers
func Setup(_ context.Context, r *mux.Router, cfg *config.Config, graphDB IGraph, datasetClient IDatasetClient, cantabularClient CantabularClient, permissions IAuthHandler, enableURLRewriting bool, observationAPIURL *url.URL) *API {
	api := &API{
		cfg:                cfg,
		Router:             r,
		graphDB:            graphDB,
		datasetClient:      datasetClient,
		cantabularClient:   cantabularClient,
		permissions:        permissions,
		enableURLRewriting: enableURLRewriting,
		observationAPIURL:  observationAPIURL,
	}

	if api.cfg.EnablePrivateEndpoints {
		read := auth.Permissions{Read: true}
		r.HandleFunc("/datasets/{dataset_id}/editions/{edition}/versions/{version}/observations", permissions.Require(read, api.getObservations)).Methods(http.MethodGet)
	} else {
		r.HandleFunc("/datasets/{dataset_id}/editions/{edition}/versions/{version}/observations", api.getObservations).Methods(http.MethodGet)
	}

	return api
}

func (api *API) checkIfAuthorised(r *http.Request, logData log.Data) (authorised bool) {
	callerIdentity := request.Caller(r.Context())
	if callerIdentity != "" {
		logData["caller_identity"] = callerIdentity
		authorised = true
	}

	userIdentity := request.User(r.Context())
	if userIdentity != "" {
		logData["user_identity"] = userIdentity
		authorised = true
	}

	logData["authenticated"] = authorised

	return authorised
}

func setJSONContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

// Close is called during graceful shutdown to give the API an opportunity to perform any required disposal task
func (*API) Close(ctx context.Context) error {
	log.Info(ctx, "graceful shutdown of api complete")
	return nil
}
