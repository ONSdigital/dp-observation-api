package api

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-observation-api/config"
	"github.com/ONSdigital/dp-observation-api/store"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

//API provides a struct to wrap the api around
type API struct {
	Router               *mux.Router
	dataStore            store.DataStore
	EnablePrePublishView bool
}

// Setup creates the API struct and its endpoints with corresponding handlers
func Setup(ctx context.Context, r *mux.Router, cfg *config.Config, dataStore store.DataStore) *API {
	api := &API{
		Router:               r,
		dataStore:            dataStore,
		EnablePrePublishView: cfg.EnablePrivateEnpoints,
	}

	r.HandleFunc("/datasets/{dataset_id}/editions/{edition}/versions/{version}/observations", api.getObservations).Methods(http.MethodGet)
	return api
}

func (api *API) authenticate(r *http.Request, logData log.Data) bool {
	// TODO we should call the authentication/authorisation library here
	return false
}

func setJSONContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

// Close closes any API dependencies
func (*API) Close(ctx context.Context) error {
	log.Event(ctx, "graceful shutdown of api complete", log.INFO)
	return nil
}
