package api

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-observation-api/store"
	"github.com/ONSdigital/log.go/log"
	"github.com/gorilla/mux"
)

//API provides a struct to wrap the api around
type API struct {
	Router    *mux.Router
	dataStore store.DataStore
}

// Setup creates the API struct and its endpoints with corresponding handlers
func Setup(ctx context.Context, r *mux.Router, dataStore store.DataStore) *API {
	api := &API{
		Router:    r,
		dataStore: dataStore,
	}

	r.HandleFunc("/datasets/{dataset_id}/editions/{edition}/versions/{version}/observations", api.getObservations).Methods(http.MethodGet)
	r.HandleFunc("/hello", HelloHandler()).Methods("GET") // TODO remove this endpoint
	return api
}

func setJSONContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
}

// Close closes any API dependencies
func (*API) Close(ctx context.Context) error {
	log.Event(ctx, "graceful shutdown of api complete", log.INFO)
	return nil
}
