package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ONSdigital/dp-api-clients-go/dataset"
	"github.com/ONSdigital/dp-graph/v2/observation"
	observationtest "github.com/ONSdigital/dp-graph/v2/observation/observationtest"
	"github.com/ONSdigital/dp-observation-api/api"
	"github.com/ONSdigital/dp-observation-api/api/mock"
	errs "github.com/ONSdigital/dp-observation-api/apierrors"
	"github.com/ONSdigital/dp-observation-api/config"
	"github.com/ONSdigital/log.go/log"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	dimension1 = dataset.VersionDimension{Name: "aggregate"}
	dimension2 = dataset.VersionDimension{Name: "geography"}
	dimension3 = dataset.VersionDimension{Name: "time"}
	dimension4 = dataset.VersionDimension{Name: "age"}
)

func TestGetObservationsReturnsOK(t *testing.T) {
	ctx := context.Background()

	t.Parallel()
	Convey("Given a request to get a single observation for a version of a dataset returns 200 OK response", t, func() {

		dimensions := []dataset.VersionDimension{
			{
				Name: "aggregate",
				URL:  "http://localhost:8081/code-lists/cpih1dim1aggid",
			},
			{
				Name: "geography",
				URL:  "http://localhost:8081/code-lists/uk-only",
			},
			{
				Name: "time",
				URL:  "http://localhost:8081/code-lists/time",
			},
		}
		usagesNotes := &[]dataset.UsageNote{{Title: "data_marking", Note: "this marks the observation with a special character"}}

		count := 0
		mockRowReader := &observationtest.StreamRowReaderMock{
			ReadFunc: func() (string, error) {
				count++
				if count == 1 {
					return "v4_2,data_marking,confidence_interval,time,time,geography_code,geography,aggregate_code,aggregate", nil
				} else if count == 2 {
					return "146.3,p,2,Month,Aug-16,K02000001,,cpi1dim1G10100,01.1 Food", nil
				}
				return "", io.EOF
			},
			CloseFunc: func(context.Context) error {
				return nil
			},
		}

		graphDbMock := &mock.IGraphMock{
			StreamCSVRowsFunc: func(ctx context.Context, instanceID string, filterID string, filters *observation.DimensionFilters, limit *int) (observation.StreamRowReader, error) {
				return mockRowReader, nil
			},
		}

		dcMock := &mock.IDatasetClientMock{
			GetFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string) (dataset.DatasetDetails, error) {
				return dataset.DatasetDetails{
					State:      dataset.StatePublished.String(),
					UsageNotes: usagesNotes,
				}, nil
			},
			GetVersionFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceAuthToken string, collectionID string, datasetID string, edition string, version string) (dataset.Version, error) {
				return dataset.Version{
					Dimensions: dimensions,
					CSVHeader:  []string{"v4_2", "data_marking", "confidence_interval", "aggregate_code", "aggregate", "geography_code", "geography", "time", "time"},
					Links: dataset.Links{
						Version: dataset.Link{
							URL: "http://localhost:8080/datasets/cpih012/editions/2017/versions/1",
							ID:  "1",
						},
					},
					State: dataset.StatePublished.String(),
				}, nil
			},
		}

		cfg, err := config.Get()
		So(err, ShouldBeNil)
		api := GetAPIWithMocks(graphDbMock, dcMock, cfg)

		Convey("When request contains query parameters where the dimension name is in lower casing", func() {
			r := httptest.NewRequest("GET", "http://localhost:8080/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
			w := httptest.NewRecorder()
			api.Router.ServeHTTP(w, r)

			So(w.Code, ShouldEqual, http.StatusOK)
			So(w.Body.String(), ShouldContainSubstring, getTestData(ctx, "expectedDocWithSingleObservation"))

			So(len(dcMock.GetCalls()), ShouldEqual, 1)
			So(len(dcMock.GetVersionCalls()), ShouldEqual, 1)
			So(len(graphDbMock.StreamCSVRowsCalls()), ShouldEqual, 1)
			So(len(mockRowReader.ReadCalls()), ShouldEqual, 3)
		})

		Convey("When request contains query parameters where the dimension name is in upper casing", func() {
			r := httptest.NewRequest("GET", "http://localhost:8080/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&AggregaTe=cpi1dim1S40403&GEOGRAPHY=K02000001", nil)
			w := httptest.NewRecorder()
			api.Router.ServeHTTP(w, r)

			So(w.Code, ShouldEqual, http.StatusOK)
			So(w.Body.String(), ShouldContainSubstring, getTestData(ctx, "expectedSecondDocWithSingleObservation"))

			So(len(dcMock.GetCalls()), ShouldEqual, 1)
			So(len(dcMock.GetVersionCalls()), ShouldEqual, 1)
			So(len(graphDbMock.StreamCSVRowsCalls()), ShouldEqual, 1)
			So(len(mockRowReader.ReadCalls()), ShouldEqual, 3)
		})

	})

	Convey("A successful request to get multiple observations via a wildcard for a version of a dataset returns 200 OK response", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:8080/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=*&geography=K02000001", nil)
		w := httptest.NewRecorder()

		dimensions := []dataset.VersionDimension{
			{
				Name: "aggregate",
				URL:  "http://localhost:8081/code-lists/cpih1dim1aggid",
			},
			{
				Name: "geography",
				URL:  "http://localhost:8081/code-lists/uk-only",
			},
			{
				Name: "time",
				URL:  "http://localhost:8081/code-lists/time",
			},
		}
		usagesNotes := &[]dataset.UsageNote{{Title: "data_marking", Note: "this marks the observation with a special character"}}

		count := 0
		mockRowReader := &observationtest.StreamRowReaderMock{
			ReadFunc: func() (string, error) {
				count++
				if count == 1 {
					return "v4_2,data_marking,confidence_interval,time,time,geography_code,geography,aggregate_code,aggregate", nil
				} else if count == 2 {
					return "146.3,p,2,Month,Aug-16,K02000001,,cpi1dim1G10100,01.1 Food", nil
				} else if count == 3 {
					return "112.1,,,Month,Aug-16,K02000001,,cpi1dim1G10101,01.2 Waste", nil
				}
				return "", io.EOF
			},
			CloseFunc: func(context.Context) error {
				return nil
			},
		}

		dcMock := &mock.IDatasetClientMock{
			GetFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string) (dataset.DatasetDetails, error) {
				return dataset.DatasetDetails{
					State:      dataset.StatePublished.String(),
					UsageNotes: usagesNotes,
				}, nil
			},
			GetVersionFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceAuthToken string, collectionID string, datasetID string, edition string, version string) (dataset.Version, error) {
				return dataset.Version{
					Dimensions: dimensions,
					CSVHeader:  []string{"v4_2", "data_marking", "confidence_interval", "aggregate_code", "aggregate", "geography_code", "geography", "time", "time"},
					Links: dataset.Links{
						Version: dataset.Link{
							URL: "http://localhost:8080/datasets/cpih012/editions/2017/versions/1",
							ID:  "1",
						},
					},
					State: dataset.StatePublished.String(),
				}, nil
			},
		}

		graphDbMock := &mock.IGraphMock{
			StreamCSVRowsFunc: func(ctx context.Context, instanceID string, filterID string, filters *observation.DimensionFilters, limit *int) (observation.StreamRowReader, error) {
				return mockRowReader, nil
			},
		}

		cfg, err := config.Get()
		So(err, ShouldBeNil)
		api := GetAPIWithMocks(graphDbMock, dcMock, cfg)
		api.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusOK)
		So(w.Body.String(), ShouldContainSubstring, getTestData(ctx, "expectedDocWithMultipleObservations"))

		So(len(dcMock.GetCalls()), ShouldEqual, 1)
		So(len(dcMock.GetVersionCalls()), ShouldEqual, 1)
		So(len(graphDbMock.StreamCSVRowsCalls()), ShouldEqual, 1)
		So(len(mockRowReader.ReadCalls()), ShouldEqual, 4)
	})
}

func TestGetObservationsReturnsError(t *testing.T) {
	t.Parallel()
	Convey("When the api cannot connect to dataset api return an internal server error", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
		w := httptest.NewRecorder()

		dcMock := &mock.IDatasetClientMock{
			GetFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string) (dataset.DatasetDetails, error) {
				return dataset.DatasetDetails{}, errs.ErrInternalServer
			},
		}

		cfg, err := config.Get()
		So(err, ShouldBeNil)
		api := GetAPIWithMocks(&mock.IGraphMock{}, dcMock, cfg)
		api.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(w.Body.String(), ShouldResemble, "internal error\n")

		So(len(dcMock.GetCalls()), ShouldEqual, 1)
	})

	Convey("When the dataset does not exist return status not found", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
		w := httptest.NewRecorder()

		dcMock := &mock.IDatasetClientMock{
			GetFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string) (dataset.DatasetDetails, error) {
				return dataset.DatasetDetails{}, errs.ErrDatasetNotFound
			},
		}

		cfg, err := config.Get()
		So(err, ShouldBeNil)
		api := GetAPIWithMocks(&mock.IGraphMock{}, dcMock, cfg)
		api.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(w.Body.String(), ShouldContainSubstring, errs.ErrDatasetNotFound.Error())

		So(len(dcMock.GetCalls()), ShouldEqual, 1)
	})

	Convey("When the dataset version does not exist return status not found", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
		w := httptest.NewRecorder()

		dcMock := &mock.IDatasetClientMock{
			GetFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string) (dataset.DatasetDetails, error) {
				return dataset.DatasetDetails{State: dataset.StatePublished.String()}, nil
			},
			GetVersionFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceAuthToken string, collectionID string, datasetID string, edition string, version string) (dataset.Version, error) {
				return dataset.Version{}, errs.ErrDatasetNotFound
			},
		}

		cfg, err := config.Get()
		So(err, ShouldBeNil)
		api := GetAPIWithMocks(&mock.IGraphMock{}, dcMock, cfg)
		api.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(w.Body.String(), ShouldContainSubstring, errs.ErrDatasetNotFound.Error())

		So(len(dcMock.GetCalls()), ShouldEqual, 1)
		So(len(dcMock.GetVersionCalls()), ShouldEqual, 1)
	})

	Convey("When the dataset exists but is unpublished return status not found for unauthorised users", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
		w := httptest.NewRecorder()

		dcMock := &mock.IDatasetClientMock{
			GetFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string) (dataset.DatasetDetails, error) {
				return dataset.DatasetDetails{State: dataset.StateCreated.String()}, nil
			},
		}

		cfg, err := config.Get()
		So(err, ShouldBeNil)
		api := GetAPIWithMocks(&mock.IGraphMock{}, dcMock, cfg)
		api.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(w.Body.String(), ShouldContainSubstring, errs.ErrDatasetNotFound.Error())

		So(len(dcMock.GetCalls()), ShouldEqual, 1)
	})

	Convey("When an unpublished version has an incorrect state for an edition of a dataset return not found error for unauthorised users", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
		w := httptest.NewRecorder()

		dcMock := &mock.IDatasetClientMock{
			GetFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string) (dataset.DatasetDetails, error) {
				return dataset.DatasetDetails{State: dataset.StatePublished.String()}, nil
			},
			GetVersionFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceAuthToken string, collectionID string, datasetID string, edition string, version string) (dataset.Version, error) {
				return dataset.Version{State: "gobbly-gook"}, nil
			},
		}

		cfg, err := config.Get()
		So(err, ShouldBeNil)
		api := GetAPIWithMocks(&mock.IGraphMock{}, dcMock, cfg)
		api.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(w.Body.String(), ShouldContainSubstring, errs.ErrDatasetNotFound.Error())

		So(len(dcMock.GetCalls()), ShouldEqual, 1)
		So(len(dcMock.GetVersionCalls()), ShouldEqual, 1)
	})

	Convey("When an unpublished version has an incorrect state return an internal error for authorised users", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
		w := httptest.NewRecorder()

		dcMock := &mock.IDatasetClientMock{
			GetFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string) (dataset.DatasetDetails, error) {
				return dataset.DatasetDetails{State: dataset.StateCreated.String()}, nil
			},
			GetVersionFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceAuthToken string, collectionID string, datasetID string, edition string, version string) (dataset.Version, error) {
				return dataset.Version{State: "gobbly-gook"}, nil
			},
		}

		cfg, err := config.Get()
		So(err, ShouldBeNil)
		cfg.EnablePrivateEndpoints = true
		api := GetAPIWithMocks(&mock.IGraphMock{}, dcMock, cfg)
		api.Router.ServeHTTP(w, r)

		assertInternalServerErr(w)
		So(len(dcMock.GetCalls()), ShouldEqual, 1)
		So(len(dcMock.GetVersionCalls()), ShouldEqual, 1)
	})

	Convey("When a version document has not got a headers field return an internal server error", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
		w := httptest.NewRecorder()

		dcMock := &mock.IDatasetClientMock{
			GetFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string) (dataset.DatasetDetails, error) {
				return dataset.DatasetDetails{State: dataset.StatePublished.String()}, nil
			},
			GetVersionFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceAuthToken string, collectionID string, datasetID string, edition string, version string) (dataset.Version, error) {
				return dataset.Version{
					Dimensions: []dataset.VersionDimension{dimension1, dimension2, dimension3},
					State:      dataset.StatePublished.String(),
				}, nil
			},
		}

		cfg, err := config.Get()
		So(err, ShouldBeNil)
		api := GetAPIWithMocks(&mock.IGraphMock{}, dcMock, cfg)
		api.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		So(len(dcMock.GetCalls()), ShouldEqual, 1)
		So(len(dcMock.GetVersionCalls()), ShouldEqual, 1)
	})

	Convey("When a version document has not got any dimensions field return an internal server error", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
		w := httptest.NewRecorder()

		dcMock := &mock.IDatasetClientMock{
			GetFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string) (dataset.DatasetDetails, error) {
				return dataset.DatasetDetails{State: dataset.StatePublished.String()}, nil
			},
			GetVersionFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceAuthToken string, collectionID string, datasetID string, edition string, version string) (dataset.Version, error) {
				return dataset.Version{
					CSVHeader: []string{"v4_0", "time_code", "time", "aggregate_code", "aggregate"},
					State:     dataset.StatePublished.String(),
				}, nil
			},
		}

		cfg, err := config.Get()
		So(err, ShouldBeNil)
		api := GetAPIWithMocks(&mock.IGraphMock{}, dcMock, cfg)
		api.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		So(len(dcMock.GetCalls()), ShouldEqual, 1)
		So(len(dcMock.GetVersionCalls()), ShouldEqual, 1)
	})

	Convey("When the first header in array does not describe the header row correctly return internal error", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
		w := httptest.NewRecorder()

		dcMock := &mock.IDatasetClientMock{
			GetFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string) (dataset.DatasetDetails, error) {
				return dataset.DatasetDetails{State: dataset.StatePublished.String()}, nil
			},
			GetVersionFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceAuthToken string, collectionID string, datasetID string, edition string, version string) (dataset.Version, error) {
				return dataset.Version{
					Dimensions: []dataset.VersionDimension{dimension1, dimension2, dimension3},
					CSVHeader:  []string{"v4"},
					State:      dataset.StatePublished.String(),
				}, nil
			},
		}

		cfg, err := config.Get()
		So(err, ShouldBeNil)
		api := GetAPIWithMocks(&mock.IGraphMock{}, dcMock, cfg)
		api.Router.ServeHTTP(w, r)

		assertInternalServerErr(w)
		So(len(dcMock.GetCalls()), ShouldEqual, 1)
		So(len(dcMock.GetVersionCalls()), ShouldEqual, 1)
	})

	Convey("When an invalid query parameter is set in request return 400 bad request with an error message containing a list of incorrect query parameters", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
		w := httptest.NewRecorder()

		dcMock := &mock.IDatasetClientMock{
			GetFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string) (dataset.DatasetDetails, error) {
				return dataset.DatasetDetails{State: dataset.StatePublished.String()}, nil
			},
			GetVersionFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceAuthToken string, collectionID string, datasetID string, edition string, version string) (dataset.Version, error) {
				return dataset.Version{
					Dimensions: []dataset.VersionDimension{dimension1, dimension3},
					CSVHeader:  []string{"v4_0", "time_code", "time", "aggregate_code", "aggregate"},
					State:      dataset.StatePublished.String(),
				}, nil
			},
		}

		cfg, err := config.Get()
		So(err, ShouldBeNil)
		api := GetAPIWithMocks(&mock.IGraphMock{}, dcMock, cfg)
		api.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Body.String(), ShouldResemble, "incorrect selection of query parameters: [geography], these dimensions do not exist for this version of the dataset\n")

		So(len(dcMock.GetCalls()), ShouldEqual, 1)
		So(len(dcMock.GetVersionCalls()), ShouldEqual, 1)
	})

	Convey("When there is a missing query parameter that is expected to be set in request return 400 bad request with an error message containing a list of missing query parameters", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
		w := httptest.NewRecorder()

		dcMock := &mock.IDatasetClientMock{
			GetFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string) (dataset.DatasetDetails, error) {
				return dataset.DatasetDetails{State: dataset.StatePublished.String()}, nil
			},
			GetVersionFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceAuthToken string, collectionID string, datasetID string, edition string, version string) (dataset.Version, error) {
				return dataset.Version{
					Dimensions: []dataset.VersionDimension{dimension1, dimension2, dimension3, dimension4},
					CSVHeader:  []string{"v4_0", "time_code", "time", "aggregate_code", "aggregate", "geography_code", "geography", "age_code", "age"},
					State:      dataset.StatePublished.String(),
				}, nil
			},
		}
		cfg, err := config.Get()
		So(err, ShouldBeNil)
		api := GetAPIWithMocks(&mock.IGraphMock{}, dcMock, cfg)
		api.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Body.String(), ShouldResemble, "missing query parameters for the following dimensions: [age]\n")

		So(len(dcMock.GetCalls()), ShouldEqual, 1)
		So(len(dcMock.GetVersionCalls()), ShouldEqual, 1)
	})

	Convey("When there are too many query parameters that are set to wildcard (*) value request returns 400 bad request", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=*&aggregate=*&geography=K02000001", nil)
		w := httptest.NewRecorder()

		dcMock := &mock.IDatasetClientMock{
			GetFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string) (dataset.DatasetDetails, error) {
				return dataset.DatasetDetails{State: dataset.StatePublished.String()}, nil
			},
			GetVersionFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceAuthToken string, collectionID string, datasetID string, edition string, version string) (dataset.Version, error) {
				return dataset.Version{
					Dimensions: []dataset.VersionDimension{dimension1, dimension2, dimension3},
					CSVHeader:  []string{"v4_0", "time_code", "time", "aggregate_code", "aggregate", "geography_code", "geography"},
					State:      dataset.StatePublished.String(),
				}, nil
			},
		}

		cfg, err := config.Get()
		So(err, ShouldBeNil)
		api := GetAPIWithMocks(&mock.IGraphMock{}, dcMock, cfg)
		api.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Body.String(), ShouldResemble, "only one wildcard (*) is allowed as a value in selected query parameters\n")

		So(len(dcMock.GetCalls()), ShouldEqual, 1)
		So(len(dcMock.GetVersionCalls()), ShouldEqual, 1)
	})

	Convey("When requested query does not find a unique observation return no observations found", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
		w := httptest.NewRecorder()

		dcMock := &mock.IDatasetClientMock{
			GetFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string) (dataset.DatasetDetails, error) {
				return dataset.DatasetDetails{State: dataset.StatePublished.String()}, nil
			},
			GetVersionFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceAuthToken string, collectionID string, datasetID string, edition string, version string) (dataset.Version, error) {
				return dataset.Version{
					Dimensions: []dataset.VersionDimension{dimension1, dimension2, dimension3},
					CSVHeader:  []string{"v4_0", "time_code", "time", "aggregate_code", "aggregate", "geography_code", "geography"},
					State:      dataset.StatePublished.String(),
				}, nil
			},
		}

		graphDBMock := &mock.IGraphMock{
			StreamCSVRowsFunc: func(context.Context, string, string, *observation.DimensionFilters, *int) (observation.StreamRowReader, error) {
				return nil, errs.ErrObservationsNotFound
			},
		}

		cfg, err := config.Get()
		So(err, ShouldBeNil)
		api := GetAPIWithMocks(graphDBMock, dcMock, cfg)
		api.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(w.Body.String(), ShouldContainSubstring, errs.ErrObservationsNotFound.Error())

		So(len(dcMock.GetCalls()), ShouldEqual, 1)
		So(len(dcMock.GetVersionCalls()), ShouldEqual, 1)
		So(len(graphDBMock.StreamCSVRowsCalls()), ShouldEqual, 1)
	})

	Convey("When requested query has a multi-valued dimension return bad request", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001&geography=K02000002", nil)
		w := httptest.NewRecorder()

		dimensions := []dataset.VersionDimension{
			{
				Name: "aggregate",
				URL:  "http://localhost:8081/code-lists/cpih1dim1aggid",
			},
			{
				Name: "geography",
				URL:  "http://localhost:8081/code-lists/uk-only",
			},
			{
				Name: "time",
				URL:  "http://localhost:8081/code-lists/time",
			},
		}
		usagesNotes := &[]dataset.UsageNote{{Title: "data_marking", Note: "this marks the observation with a special character"}}

		dcMock := &mock.IDatasetClientMock{
			GetFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string) (dataset.DatasetDetails, error) {
				return dataset.DatasetDetails{
					State:      dataset.StatePublished.String(),
					UsageNotes: usagesNotes,
				}, nil
			},
			GetVersionFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceAuthToken string, collectionID string, datasetID string, edition string, version string) (dataset.Version, error) {
				return dataset.Version{
					Dimensions: dimensions,
					CSVHeader:  []string{"v4_2", "data_marking", "confidence_interval", "aggregate_code", "aggregate", "geography_code", "geography", "time", "time"},
					Links: dataset.Links{
						Version: dataset.Link{
							URL: "http://localhost:8080/datasets/cpih012/editions/2017/versions/1",
							ID:  "1",
						},
					},
					State: dataset.StatePublished.String(),
				}, nil
			},
		}

		cfg, err := config.Get()
		So(err, ShouldBeNil)
		api := GetAPIWithMocks(&mock.IGraphMock{}, dcMock, cfg)
		api.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Body.String(), ShouldResemble, "multi-valued query parameters for the following dimensions: [geography]\n")

		So(len(dcMock.GetVersionCalls()), ShouldEqual, 1)
	})
}

func TestGetListOfValidDimensionNames(t *testing.T) {
	t.Parallel()
	Convey("Given a list of valid dimension codelist objects", t, func() {
		Convey("When getListOfValidDimensionNames is called", func() {
			dimension1 := dataset.VersionDimension{
				Name: "time",
			}

			dimension2 := dataset.VersionDimension{
				Name: "aggregate",
			}

			dimension3 := dataset.VersionDimension{
				Name: "geography",
			}

			version := &dataset.Version{
				Dimensions: []dataset.VersionDimension{dimension1, dimension2, dimension3},
			}

			Convey("Then func returns the correct number of dimensions", func() {
				validDimensions := api.GetListOfValidDimensionNames(version.Dimensions)

				So(len(validDimensions), ShouldEqual, 3)
				So(validDimensions[0], ShouldEqual, "time")
				So(validDimensions[1], ShouldEqual, "aggregate")
				So(validDimensions[2], ShouldEqual, "geography")
			})
		})
	})
}

func TestGetDimensionOffsetInHeaderRow(t *testing.T) {
	t.Parallel()
	Convey("Given the version headers are valid", t, func() {
		Convey("When the version has no metadata headers", func() {
			version := &dataset.Version{
				CSVHeader: []string{
					"v4_0",
					"time_codelist",
					"time",
					"aggregate_codelist",
					"Aggregate",
					"geography_codelist",
					"geography",
				},
			}

			Convey("Then getListOfValidDimensionNames func returns the correct number of headers", func() {
				dimensionOffset, err := api.GetDimensionOffsetInHeaderRow(version.CSVHeader)

				So(err, ShouldBeNil)
				So(dimensionOffset, ShouldEqual, 0)
			})
		})

		Convey("When the version has metadata headers", func() {
			version := &dataset.Version{
				CSVHeader: []string{
					"V4_2",
					"data_marking",
					"confidence_interval",
					"time_codelist",
					"time",
				},
			}

			Convey("Then getListOfValidDimensionNames func returns the correct number of headers", func() {
				dimensionOffset, err := api.GetDimensionOffsetInHeaderRow(version.CSVHeader)

				So(err, ShouldBeNil)
				So(dimensionOffset, ShouldEqual, 2)
			})
		})
	})

	Convey("Given the first value in the header does not have an underscore `_` in value", t, func() {
		Convey("When the getListOfValidDimensionNames func is called", func() {
			version := &dataset.Version{
				CSVHeader: []string{
					"v4",
					"time_codelist",
					"time",
					"aggregate_codelist",
					"aggregate",
					"geography_codelist",
					"geography",
				},
			}
			Convey("Then function returns error, `index out of range`", func() {
				dimensionOffset, err := api.GetDimensionOffsetInHeaderRow(version.CSVHeader)

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldResemble, "index out of range")
				So(dimensionOffset, ShouldEqual, 0)
			})
		})
	})

	Convey("Given the first value in the header does not follow the format `v4_1`", t, func() {
		Convey("When the getListOfValidDimensionNames func is called", func() {
			version := &dataset.Version{
				CSVHeader: []string{
					"v4_one",
					"time_codelist",
					"time",
					"aggregate_codelist",
					"aggregate",
					"geography_codelist",
					"geography",
				},
			}
			Convey("Then function returns error, `index out of range`", func() {
				dimensionOffset, err := api.GetDimensionOffsetInHeaderRow(version.CSVHeader)

				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldResemble, "strconv.Atoi: parsing \"one\": invalid syntax")
				So(dimensionOffset, ShouldEqual, 0)
			})
		})
	})
}

func TestExtractQueryParameters(t *testing.T) {
	t.Parallel()
	Convey("Given a list of valid dimension headers for version", t, func() {
		headers := []string{
			"time",
			"aggregate",
			"geography",
		}

		Convey("When a request is made containing query parameters for each dimension/header", func() {
			r, err := http.NewRequest("GET",
				"http://localhost:22000/datasets/123/editions/2017/versions/1/observations?time=JAN08&aggregate=Overall Index&geography=wales",
				nil,
			)
			So(err, ShouldBeNil)

			Convey("Then extractQueryParameters func returns a list of query parameters and their corresponding value", func() {
				queryParameters, err := api.ExtractQueryParameters(r.URL.Query(), headers)
				So(err, ShouldBeNil)
				So(len(queryParameters), ShouldEqual, 3)
				So(queryParameters["time"], ShouldEqual, "JAN08")
				So(queryParameters["aggregate"], ShouldEqual, "Overall Index")
				So(queryParameters["geography"], ShouldEqual, "wales")
			})
		})

		Convey("When a request is made containing query parameters for 2/3 dimensions/headers", func() {
			r, err := http.NewRequest("GET",
				"http://localhost:22000/datasets/123/editions/2017/versions/1/observations?time=JAN08&geography=wales",
				nil,
			)
			So(err, ShouldBeNil)

			Convey("Then extractQueryParameters func returns an error", func() {
				queryParameters, err := api.ExtractQueryParameters(r.URL.Query(), headers)
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, errs.ErrorMissingQueryParameters([]string{"aggregate"}))
				So(queryParameters, ShouldBeNil)
			})
		})

		Convey("When a request is made containing all query parameters for each dimensions/headers but also an invalid one", func() {
			r, err := http.NewRequest("GET",
				"http://localhost:22000/datasets/123/editions/2017/versions/1/observations?time=JAN08&aggregate=Food&geography=wales&age=52",
				nil,
			)
			So(err, ShouldBeNil)

			Convey("Then extractQueryParameters func returns an error", func() {
				queryParameters, err := api.ExtractQueryParameters(r.URL.Query(), headers)
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, errs.ErrorIncorrectQueryParameters([]string{"age"}))
				So(queryParameters, ShouldBeNil)
			})
		})

		Convey("When a request is made containing all query parameters for each dimensions/headers but there is a duplicate", func() {
			r, err := http.NewRequest("GET",
				"http://localhost:22000/datasets/123/editions/2017/versions/1/observations?time=JAN08&aggregate=Food&geography=wales&time=JAN0",
				nil,
			)
			So(err, ShouldBeNil)

			Convey("Then extractQueryParameters func returns an error", func() {
				queryParameters, err := api.ExtractQueryParameters(r.URL.Query(), headers)
				So(err, ShouldNotBeNil)
				So(err, ShouldResemble, errs.ErrorMultivaluedQueryParameters([]string{"time"}))
				So(queryParameters, ShouldBeNil)
			})
		})
	})
}

func getTestData(ctx context.Context, filename string) string {
	jsonBytes, err := ioutil.ReadFile("./observation_test_data/" + filename + ".json")
	if err != nil {
		log.Event(ctx, "unable to read json file into bytes", log.ERROR, log.Error(err), log.Data{"filename": filename})
		os.Exit(1)
	}
	buffer := new(bytes.Buffer)
	if err := json.Compact(buffer, jsonBytes); err != nil {
		log.Event(ctx, "unable to remove whitespace from json bytes", log.ERROR, log.Error(err), log.Data{"filename": filename})
		os.Exit(1)
	}

	return buffer.String()
}
