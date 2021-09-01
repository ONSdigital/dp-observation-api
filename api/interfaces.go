package api

import (
	"context"
	"net/http"

	dataset "github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-authorisation/auth"
	"github.com/ONSdigital/dp-graph/v2/graph/driver"
	"github.com/ONSdigital/dp-graph/v2/observation"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
)

//go:generate moq -out mock/graph.go -pkg mock . IGraph
//go:generate moq -out mock/dataset.go -pkg mock . IDatasetClient
//go:generate moq -out mock/authorisation.go -pkg mock . IAuthHandler

// IGraph defines the required methods from GraphDB required by Observation API
type IGraph interface {
	driver.Driver
	StreamCSVRows(ctx context.Context, instanceID, filterID string, filters *observation.DimensionFilters, limit *int) (observation.StreamRowReader, error)
	ErrorChan() chan error
}

// IDatasetClient represents the required methods from the Dataset Client required by Observation API
type IDatasetClient interface {
	GetVersion(ctx context.Context, userAuthToken, serviceAuthToken, downloadServiceAuthToken, collectionID, datasetID, edition, version string) (m dataset.Version, err error)
	Get(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, datasetID string) (m dataset.DatasetDetails, err error)
	Checker(ctx context.Context, check *healthcheck.CheckState) error
	GetOptions(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, edition, version, dimension string, q *dataset.QueryParams) (m dataset.Options, err error)
}

// IAuthHandler represents the required methods from authorisation library by Observation API
type IAuthHandler interface {
	Require(required auth.Permissions, handler http.HandlerFunc) http.HandlerFunc
}
