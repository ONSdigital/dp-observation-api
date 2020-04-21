// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package storetest

import (
	"context"
	"github.com/ONSdigital/dp-graph/v2/observation"
	"github.com/ONSdigital/dp-observation-api/models"
	"github.com/ONSdigital/dp-observation-api/store"
	"sync"
)

var (
	lockStorerMockCheckEditionExists sync.RWMutex
	lockStorerMockGetDataset         sync.RWMutex
	lockStorerMockGetVersion         sync.RWMutex
	lockStorerMockStreamCSVRows      sync.RWMutex
)

// Ensure, that StorerMock does implement store.Storer.
// If this is not the case, regenerate this file with moq.
var _ store.Storer = &StorerMock{}

// StorerMock is a mock implementation of store.Storer.
//
//     func TestSomethingThatUsesStorer(t *testing.T) {
//
//         // make and configure a mocked store.Storer
//         mockedStorer := &StorerMock{
//             CheckEditionExistsFunc: func(ID string, editionID string, state string) error {
// 	               panic("mock out the CheckEditionExists method")
//             },
//             GetDatasetFunc: func(ID string) (*models.DatasetUpdate, error) {
// 	               panic("mock out the GetDataset method")
//             },
//             GetVersionFunc: func(datasetID string, editionID string, version string, state string) (*models.Version, error) {
// 	               panic("mock out the GetVersion method")
//             },
//             StreamCSVRowsFunc: func(ctx context.Context, instanceID string, filterID string, filters *observation.DimensionFilters, limit *int) (observation.StreamRowReader, error) {
// 	               panic("mock out the StreamCSVRows method")
//             },
//         }
//
//         // use mockedStorer in code that requires store.Storer
//         // and then make assertions.
//
//     }
type StorerMock struct {
	// CheckEditionExistsFunc mocks the CheckEditionExists method.
	CheckEditionExistsFunc func(ID string, editionID string, state string) error

	// GetDatasetFunc mocks the GetDataset method.
	GetDatasetFunc func(ID string) (*models.DatasetUpdate, error)

	// GetVersionFunc mocks the GetVersion method.
	GetVersionFunc func(datasetID string, editionID string, version string, state string) (*models.Version, error)

	// StreamCSVRowsFunc mocks the StreamCSVRows method.
	StreamCSVRowsFunc func(ctx context.Context, instanceID string, filterID string, filters *observation.DimensionFilters, limit *int) (observation.StreamRowReader, error)

	// calls tracks calls to the methods.
	calls struct {
		// CheckEditionExists holds details about calls to the CheckEditionExists method.
		CheckEditionExists []struct {
			// ID is the ID argument value.
			ID string
			// EditionID is the editionID argument value.
			EditionID string
			// State is the state argument value.
			State string
		}
		// GetDataset holds details about calls to the GetDataset method.
		GetDataset []struct {
			// ID is the ID argument value.
			ID string
		}
		// GetVersion holds details about calls to the GetVersion method.
		GetVersion []struct {
			// DatasetID is the datasetID argument value.
			DatasetID string
			// EditionID is the editionID argument value.
			EditionID string
			// Version is the version argument value.
			Version string
			// State is the state argument value.
			State string
		}
		// StreamCSVRows holds details about calls to the StreamCSVRows method.
		StreamCSVRows []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// InstanceID is the instanceID argument value.
			InstanceID string
			// FilterID is the filterID argument value.
			FilterID string
			// Filters is the filters argument value.
			Filters *observation.DimensionFilters
			// Limit is the limit argument value.
			Limit *int
		}
	}
}

// CheckEditionExists calls CheckEditionExistsFunc.
func (mock *StorerMock) CheckEditionExists(ID string, editionID string, state string) error {
	if mock.CheckEditionExistsFunc == nil {
		panic("StorerMock.CheckEditionExistsFunc: method is nil but Storer.CheckEditionExists was just called")
	}
	callInfo := struct {
		ID        string
		EditionID string
		State     string
	}{
		ID:        ID,
		EditionID: editionID,
		State:     state,
	}
	lockStorerMockCheckEditionExists.Lock()
	mock.calls.CheckEditionExists = append(mock.calls.CheckEditionExists, callInfo)
	lockStorerMockCheckEditionExists.Unlock()
	return mock.CheckEditionExistsFunc(ID, editionID, state)
}

// CheckEditionExistsCalls gets all the calls that were made to CheckEditionExists.
// Check the length with:
//     len(mockedStorer.CheckEditionExistsCalls())
func (mock *StorerMock) CheckEditionExistsCalls() []struct {
	ID        string
	EditionID string
	State     string
} {
	var calls []struct {
		ID        string
		EditionID string
		State     string
	}
	lockStorerMockCheckEditionExists.RLock()
	calls = mock.calls.CheckEditionExists
	lockStorerMockCheckEditionExists.RUnlock()
	return calls
}

// GetDataset calls GetDatasetFunc.
func (mock *StorerMock) GetDataset(ID string) (*models.DatasetUpdate, error) {
	if mock.GetDatasetFunc == nil {
		panic("StorerMock.GetDatasetFunc: method is nil but Storer.GetDataset was just called")
	}
	callInfo := struct {
		ID string
	}{
		ID: ID,
	}
	lockStorerMockGetDataset.Lock()
	mock.calls.GetDataset = append(mock.calls.GetDataset, callInfo)
	lockStorerMockGetDataset.Unlock()
	return mock.GetDatasetFunc(ID)
}

// GetDatasetCalls gets all the calls that were made to GetDataset.
// Check the length with:
//     len(mockedStorer.GetDatasetCalls())
func (mock *StorerMock) GetDatasetCalls() []struct {
	ID string
} {
	var calls []struct {
		ID string
	}
	lockStorerMockGetDataset.RLock()
	calls = mock.calls.GetDataset
	lockStorerMockGetDataset.RUnlock()
	return calls
}

// GetVersion calls GetVersionFunc.
func (mock *StorerMock) GetVersion(datasetID string, editionID string, version string, state string) (*models.Version, error) {
	if mock.GetVersionFunc == nil {
		panic("StorerMock.GetVersionFunc: method is nil but Storer.GetVersion was just called")
	}
	callInfo := struct {
		DatasetID string
		EditionID string
		Version   string
		State     string
	}{
		DatasetID: datasetID,
		EditionID: editionID,
		Version:   version,
		State:     state,
	}
	lockStorerMockGetVersion.Lock()
	mock.calls.GetVersion = append(mock.calls.GetVersion, callInfo)
	lockStorerMockGetVersion.Unlock()
	return mock.GetVersionFunc(datasetID, editionID, version, state)
}

// GetVersionCalls gets all the calls that were made to GetVersion.
// Check the length with:
//     len(mockedStorer.GetVersionCalls())
func (mock *StorerMock) GetVersionCalls() []struct {
	DatasetID string
	EditionID string
	Version   string
	State     string
} {
	var calls []struct {
		DatasetID string
		EditionID string
		Version   string
		State     string
	}
	lockStorerMockGetVersion.RLock()
	calls = mock.calls.GetVersion
	lockStorerMockGetVersion.RUnlock()
	return calls
}

// StreamCSVRows calls StreamCSVRowsFunc.
func (mock *StorerMock) StreamCSVRows(ctx context.Context, instanceID string, filterID string, filters *observation.DimensionFilters, limit *int) (observation.StreamRowReader, error) {
	if mock.StreamCSVRowsFunc == nil {
		panic("StorerMock.StreamCSVRowsFunc: method is nil but Storer.StreamCSVRows was just called")
	}
	callInfo := struct {
		Ctx        context.Context
		InstanceID string
		FilterID   string
		Filters    *observation.DimensionFilters
		Limit      *int
	}{
		Ctx:        ctx,
		InstanceID: instanceID,
		FilterID:   filterID,
		Filters:    filters,
		Limit:      limit,
	}
	lockStorerMockStreamCSVRows.Lock()
	mock.calls.StreamCSVRows = append(mock.calls.StreamCSVRows, callInfo)
	lockStorerMockStreamCSVRows.Unlock()
	return mock.StreamCSVRowsFunc(ctx, instanceID, filterID, filters, limit)
}

// StreamCSVRowsCalls gets all the calls that were made to StreamCSVRows.
// Check the length with:
//     len(mockedStorer.StreamCSVRowsCalls())
func (mock *StorerMock) StreamCSVRowsCalls() []struct {
	Ctx        context.Context
	InstanceID string
	FilterID   string
	Filters    *observation.DimensionFilters
	Limit      *int
} {
	var calls []struct {
		Ctx        context.Context
		InstanceID string
		FilterID   string
		Filters    *observation.DimensionFilters
		Limit      *int
	}
	lockStorerMockStreamCSVRows.RLock()
	calls = mock.calls.StreamCSVRows
	lockStorerMockStreamCSVRows.RUnlock()
	return calls
}
