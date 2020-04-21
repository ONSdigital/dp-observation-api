package store

import (
	"context"

	"github.com/ONSdigital/dp-graph/v2/observation"
	"github.com/ONSdigital/dp-observation-api/models"
)

// DataStore provides a datastore.Storer interface used to store, retrieve, remove or update datasets
type DataStore struct {
	Backend Storer
}

//go:generate moq -out datastoretest/datastore.go -pkg storetest . Storer

// Storer represents basic data access via Get, Remove and Upsert methods.
type Storer interface {
	GetDataset(ID string) (*models.DatasetUpdate, error)
	CheckEditionExists(ID, editionID, state string) error
	GetVersion(datasetID, editionID, version, state string) (*models.Version, error)

	StreamCSVRows(ctx context.Context, instanceID, filterID string, filters *observation.DimensionFilters, limit *int) (observation.StreamRowReader, error)
}
