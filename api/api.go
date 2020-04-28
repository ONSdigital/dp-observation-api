package api

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-observation-api/config"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

//API provides a struct to wrap the api around
type API struct {
	Router        *mux.Router
	graphDB       IGraph
	datasetClient IDatasetClient
	cfg           *config.Config
}

// Setup creates the API struct and its endpoints with corresponding handlers
func Setup(ctx context.Context, r *mux.Router, cfg *config.Config, graphDB IGraph, datasetClient IDatasetClient) *API {
	api := &API{
		Router:        r,
		graphDB:       graphDB,
		datasetClient: datasetClient,
		cfg:           cfg,
	}

	r.HandleFunc("/datasets/{dataset_id}/editions/{edition}/versions/{version}/observations", api.getObservations).Methods(http.MethodGet)
	return api
}

func (api *API) authenticate(r *http.Request, logData log.Data) bool {
	// TODO we should call the authentication/authorisation library here
	var authorised bool

	if api.cfg.EnablePrivateEndpoints {
		return true
	}
	return authorised
}

func setJSONContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

// Close closes any API dependencies
func (*API) Close(ctx context.Context) error {
	log.Event(ctx, "graceful shutdown of api complete", log.INFO)
	return nil
}
