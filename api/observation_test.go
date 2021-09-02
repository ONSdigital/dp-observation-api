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

	"github.com/ONSdigital/dp-net/request"
	"github.com/pkg/errors"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-authorisation/auth"
	"github.com/ONSdigital/dp-graph/v2/observation"
	"github.com/ONSdigital/dp-graph/v2/observation/observationtest"
	"github.com/ONSdigital/dp-observation-api/api"
	"github.com/ONSdigital/dp-observation-api/api/mock"
	errs "github.com/ONSdigital/dp-observation-api/apierrors"
	"github.com/ONSdigital/dp-observation-api/config"
	"github.com/ONSdigital/dp-observation-api/models"
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

		graphDBMock := &mock.IGraphMock{
			StreamCSVRowsFunc: func(ctx context.Context, instanceID string, filterID string, filters *observation.DimensionFilters, limit *int) (observation.StreamRowReader, error) {
				return mockRowReader, nil
			},
		}

		cMock := &mock.CantabularClientMock{}

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
					Links: dataset.Links{
						Dataset: dataset.Link{ID: "cpih012"},
						Edition: dataset.Link{ID: "2017"},
						Version: dataset.Link{
							URL: "http://localhost:8080/datasets/cpih012/editions/2017/versions/1",
							ID:  "1",
						},
					},
					State: dataset.StatePublished.String(),
				}, nil
			},
		}

		originalFunc := api.SortFilter
		defer func() {
			api.SortFilter = originalFunc
		}()
		api.SortFilter = func(ctx context.Context, api *api.API, event *models.FilterSubmitted, dbFilter *observation.DimensionFilters) {
		}

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		cfg.ObservationAPIURL = "http://localhost:8082"
		ap := GetAPIWithMocks(cfg, graphDBMock, dcMock, cMock, &auth.NopHandler{})

		Convey("When request contains query parameters where the dimension name is in lower casing", func() {
			r := httptest.NewRequest("GET", "http://localhost:8082/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
			r = r.WithContext(context.WithValue(r.Context(), request.FlorenceIdentityKey, testUserAuthToken))
			w := httptest.NewRecorder()
			ap.Router.ServeHTTP(w, r)

			So(w.Code, ShouldEqual, http.StatusOK)
			So(w.Body.String(), ShouldContainSubstring, getTestData(ctx, "expectedDocWithSingleObservation"))

			validateGetDataset(dcMock, "cpih012")
			validateGetVersion(dcMock, "cpih012", "2017", "1")
			So(len(graphDBMock.StreamCSVRowsCalls()), ShouldEqual, 1)
			So(len(mockRowReader.ReadCalls()), ShouldEqual, 3)
		})

		Convey("When request contains query parameters where the dimension name is in upper casing", func() {
			r := httptest.NewRequest("GET", "http://localhost:8080/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&AggregaTe=cpi1dim1S40403&GEOGRAPHY=K02000001", nil)
			r = r.WithContext(context.WithValue(r.Context(), request.FlorenceIdentityKey, testUserAuthToken))
			w := httptest.NewRecorder()
			ap.Router.ServeHTTP(w, r)

			So(w.Code, ShouldEqual, http.StatusOK)
			So(w.Body.String(), ShouldContainSubstring, getTestData(ctx, "expectedSecondDocWithSingleObservation"))

			validateGetDataset(dcMock, "cpih012")
			validateGetVersion(dcMock, "cpih012", "2017", "1")
			So(len(graphDBMock.StreamCSVRowsCalls()), ShouldEqual, 1)
			So(len(mockRowReader.ReadCalls()), ShouldEqual, 3)
		})
	})

	Convey("A successful request to get multiple observations via a wildcard for a version of a dataset returns 200 OK response", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:8080/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=*&geography=K02000001", nil)
		r = r.WithContext(context.WithValue(r.Context(), request.FlorenceIdentityKey, testUserAuthToken))
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
					Links: dataset.Links{
						Dataset: dataset.Link{ID: "cpih012"},
						Edition: dataset.Link{ID: "2017"},
						Version: dataset.Link{
							URL: "http://localhost:8080/datasets/cpih012/editions/2017/versions/1",
							ID:  "1",
						},
					},
					State: dataset.StatePublished.String(),
				}, nil
			},
		}

		cMock := &mock.CantabularClientMock{}

		graphDBMock := &mock.IGraphMock{
			StreamCSVRowsFunc: func(ctx context.Context, instanceID string, filterID string, filters *observation.DimensionFilters, limit *int) (observation.StreamRowReader, error) {
				return mockRowReader, nil
			},
		}

		originalFunc := api.SortFilter
		defer func() {
			api.SortFilter = originalFunc
		}()
		api.SortFilter = func(ctx context.Context, api *api.API, event *models.FilterSubmitted, dbFilter *observation.DimensionFilters) {
		}

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		cfg.ObservationAPIURL = "http://localhost:8082"

		ap := GetAPIWithMocks(cfg, graphDBMock, dcMock, cMock, &auth.NopHandler{})
		ap.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusOK)
		So(w.Body.String(), ShouldContainSubstring, getTestData(ctx, "expectedDocWithMultipleObservations"))

		validateGetDataset(dcMock, "cpih012")
		validateGetVersion(dcMock, "cpih012", "2017", "1")
		So(len(graphDBMock.StreamCSVRowsCalls()), ShouldEqual, 1)
		So(len(mockRowReader.ReadCalls()), ShouldEqual, 4)
	})
}

func TestGetObservationsReturnsError(t *testing.T) {
	t.Parallel()
	Convey("When the api cannot connect to dataset api return an internal server error", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
		r = r.WithContext(context.WithValue(r.Context(), request.FlorenceIdentityKey, testUserAuthToken))
		w := httptest.NewRecorder()

		dcMock := &mock.IDatasetClientMock{
			GetFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string) (dataset.DatasetDetails, error) {
				return dataset.DatasetDetails{}, errs.ErrInternalServer
			},
		}

		cMock := &mock.CantabularClientMock{}

		cfg, err := config.Get()
		So(err, ShouldBeNil)
		api := GetAPIWithMocks(cfg, &mock.IGraphMock{}, dcMock, cMock, &auth.NopHandler{})
		api.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusInternalServerError)
		So(w.Body.String(), ShouldResemble, "internal error\n")

		validateGetDataset(dcMock, "cpih012")
	})

	Convey("When the dataset does not exist return status not found", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
		r = r.WithContext(context.WithValue(r.Context(), request.FlorenceIdentityKey, testUserAuthToken))
		w := httptest.NewRecorder()

		dcMock := &mock.IDatasetClientMock{
			GetFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string) (dataset.DatasetDetails, error) {
				return dataset.DatasetDetails{}, errs.ErrDatasetNotFound
			},
		}

		cMock := &mock.CantabularClientMock{}

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		api := GetAPIWithMocks(cfg, &mock.IGraphMock{}, dcMock, cMock, &auth.NopHandler{})
		api.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(w.Body.String(), ShouldContainSubstring, errs.ErrDatasetNotFound.Error())

		validateGetDataset(dcMock, "cpih012")
	})

	Convey("When the dataset version does not exist return status not found", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
		r = r.WithContext(context.WithValue(r.Context(), request.FlorenceIdentityKey, testUserAuthToken))
		w := httptest.NewRecorder()

		dcMock := &mock.IDatasetClientMock{
			GetFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string) (dataset.DatasetDetails, error) {
				return dataset.DatasetDetails{State: dataset.StatePublished.String()}, nil
			},
			GetVersionFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceAuthToken string, collectionID string, datasetID string, edition string, version string) (dataset.Version, error) {
				return dataset.Version{}, errs.ErrVersionNotFound
			},
		}

		cMock := &mock.CantabularClientMock{}

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		api := GetAPIWithMocks(cfg, &mock.IGraphMock{}, dcMock, cMock, &auth.NopHandler{})
		api.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(w.Body.String(), ShouldContainSubstring, errs.ErrVersionNotFound.Error())

		validateGetDataset(dcMock, "cpih012")
		validateGetVersion(dcMock, "cpih012", "2017", "1")
	})

	Convey("When the dataset exists but is unpublished return status not found for unauthorised users", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
		r = r.WithContext(context.WithValue(r.Context(), request.FlorenceIdentityKey, testUserAuthToken))
		w := httptest.NewRecorder()

		dcMock := &mock.IDatasetClientMock{
			GetFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string) (dataset.DatasetDetails, error) {
				return dataset.DatasetDetails{State: dataset.StateCreated.String()}, nil
			},
		}

		cMock := &mock.CantabularClientMock{}

		cfg, err := config.Get()
		So(err, ShouldBeNil)
		api := GetAPIWithMocks(cfg, &mock.IGraphMock{}, dcMock, cMock, &auth.NopHandler{})
		api.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(w.Body.String(), ShouldContainSubstring, errs.ErrDatasetNotFound.Error())

		validateGetDataset(dcMock, "cpih012")
	})

	Convey("When an unpublished version has an incorrect state for an edition of a dataset return not found error for unauthorised users", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
		r = r.WithContext(context.WithValue(r.Context(), request.FlorenceIdentityKey, testUserAuthToken))
		w := httptest.NewRecorder()

		dcMock := &mock.IDatasetClientMock{
			GetFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string) (dataset.DatasetDetails, error) {
				return dataset.DatasetDetails{State: dataset.StatePublished.String()}, nil
			},
			GetVersionFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceAuthToken string, collectionID string, datasetID string, edition string, version string) (dataset.Version, error) {
				return dataset.Version{State: "gobbly-gook"}, nil
			},
		}

		cMock := &mock.CantabularClientMock{}

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		api := GetAPIWithMocks(cfg, &mock.IGraphMock{}, dcMock, cMock, &auth.NopHandler{})
		api.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(w.Body.String(), ShouldContainSubstring, errs.ErrVersionNotFound.Error())

		validateGetDataset(dcMock, "cpih012")
		validateGetVersion(dcMock, "cpih012", "2017", "1")
	})

	Convey("When an unpublished version has an incorrect state return an internal error for authorised users", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
		r = r.WithContext(context.WithValue(r.Context(), request.FlorenceIdentityKey, testUserAuthToken))
		w := httptest.NewRecorder()

		dcMock := &mock.IDatasetClientMock{
			GetFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string) (dataset.DatasetDetails, error) {
				return dataset.DatasetDetails{State: dataset.StateCreated.String()}, nil
			},
			GetVersionFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceAuthToken string, collectionID string, datasetID string, edition string, version string) (dataset.Version, error) {
				return dataset.Version{State: "gobbly-gook"}, nil
			},
		}

		cMock := &mock.CantabularClientMock{}

		pMock := &mock.IAuthHandlerMock{
			RequireFunc: func(required auth.Permissions, handler http.HandlerFunc) http.HandlerFunc {
				r = r.WithContext(context.WithValue(r.Context(), request.UserIdentityKey, testUserAuthToken))
				return func(w http.ResponseWriter, r *http.Request) {
					handler.ServeHTTP(w, r)
				}
			},
		}

		cfg, err := config.Get()
		So(err, ShouldBeNil)
		cfg.EnablePrivateEndpoints = true

		api := GetAPIWithMocks(cfg, &mock.IGraphMock{}, dcMock, cMock, pMock)
		api.Router.ServeHTTP(w, r)

		assertInternalServerErr(w)
		validateGetDataset(dcMock, "cpih012")
		validateGetVersion(dcMock, "cpih012", "2017", "1")
	})

	Convey("When graph instance node has not got a headers field return an internal server error", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
		r = r.WithContext(context.WithValue(r.Context(), request.FlorenceIdentityKey, testUserAuthToken))
		w := httptest.NewRecorder()

		mockRowReader := &observationtest.StreamRowReaderMock{
			ReadFunc: func() (string, error) {
				return "146.3,p,2,Month,Aug-16,K02000001,,cpi1dim1G10100,01.1 Food", nil
			},
			CloseFunc: func(context.Context) error {
				return nil
			},
		}

		graphDBMock := &mock.IGraphMock{
			StreamCSVRowsFunc: func(ctx context.Context, instanceID string, filterID string, filters *observation.DimensionFilters, limit *int) (observation.StreamRowReader, error) {
				return mockRowReader, nil
			},
		}

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

		cMock := &mock.CantabularClientMock{}

		originalFunc := api.SortFilter
		defer func() {
			api.SortFilter = originalFunc
		}()
		api.SortFilter = func(ctx context.Context, api *api.API, event *models.FilterSubmitted, dbFilter *observation.DimensionFilters) {
		}

		cfg, err := config.Get()
		So(err, ShouldBeNil)
		ap := GetAPIWithMocks(cfg, graphDBMock, dcMock, cMock, &auth.NopHandler{})
		ap.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		validateGetDataset(dcMock, "cpih012")
		validateGetVersion(dcMock, "cpih012", "2017", "1")
		So(len(graphDBMock.StreamCSVRowsCalls()), ShouldEqual, 1)
	})

	Convey("When a version document has not got any dimensions field return an internal server error", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
		r = r.WithContext(context.WithValue(r.Context(), request.FlorenceIdentityKey, testUserAuthToken))
		w := httptest.NewRecorder()

		dcMock := &mock.IDatasetClientMock{
			GetFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string) (dataset.DatasetDetails, error) {
				return dataset.DatasetDetails{State: dataset.StatePublished.String()}, nil
			},
			GetVersionFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceAuthToken string, collectionID string, datasetID string, edition string, version string) (dataset.Version, error) {
				return dataset.Version{
					State: dataset.StatePublished.String(),
				}, nil
			},
		}

		cMock := &mock.CantabularClientMock{}

		cfg, err := config.Get()
		So(err, ShouldBeNil)
		api := GetAPIWithMocks(cfg, &mock.IGraphMock{}, dcMock, cMock, &auth.NopHandler{})
		api.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusInternalServerError)

		validateGetDataset(dcMock, "cpih012")
		validateGetVersion(dcMock, "cpih012", "2017", "1")
	})

	Convey("When the first header in array does not describe the header row correctly return internal error", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
		r = r.WithContext(context.WithValue(r.Context(), request.FlorenceIdentityKey, testUserAuthToken))
		w := httptest.NewRecorder()

		mockRowReader := &observationtest.StreamRowReaderMock{
			ReadFunc: func() (string, error) {
				return "v4,data_marking,confidence_interval,time,time,geography_code,geography,aggregate_code,aggregate", nil
			},
			CloseFunc: func(context.Context) error {
				return nil
			},
		}

		graphDBMock := &mock.IGraphMock{
			StreamCSVRowsFunc: func(ctx context.Context, instanceID string, filterID string, filters *observation.DimensionFilters, limit *int) (observation.StreamRowReader, error) {
				return mockRowReader, nil
			},
		}

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

		cMock := &mock.CantabularClientMock{}

		originalFunc := api.SortFilter
		defer func() {
			api.SortFilter = originalFunc
		}()
		api.SortFilter = func(ctx context.Context, api *api.API, event *models.FilterSubmitted, dbFilter *observation.DimensionFilters) {
		}

		cfg, err := config.Get()
		So(err, ShouldBeNil)
		ap := GetAPIWithMocks(cfg, graphDBMock, dcMock, cMock, &auth.NopHandler{})
		ap.Router.ServeHTTP(w, r)

		assertInternalServerErr(w)
		validateGetDataset(dcMock, "cpih012")
		validateGetVersion(dcMock, "cpih012", "2017", "1")
		So(len(graphDBMock.StreamCSVRowsCalls()), ShouldEqual, 1)
	})

	Convey("When an invalid query parameter is set in request return 400 bad request with an error message containing a list of incorrect query parameters", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
		r = r.WithContext(context.WithValue(r.Context(), request.FlorenceIdentityKey, testUserAuthToken))
		w := httptest.NewRecorder()

		dcMock := &mock.IDatasetClientMock{
			GetFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string) (dataset.DatasetDetails, error) {
				return dataset.DatasetDetails{State: dataset.StatePublished.String()}, nil
			},
			GetVersionFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceAuthToken string, collectionID string, datasetID string, edition string, version string) (dataset.Version, error) {
				return dataset.Version{
					Dimensions: []dataset.VersionDimension{dimension1, dimension3},
					State:      dataset.StatePublished.String(),
				}, nil
			},
		}

		cMock := &mock.CantabularClientMock{}

		cfg, err := config.Get()
		So(err, ShouldBeNil)
		api := GetAPIWithMocks(cfg, &mock.IGraphMock{}, dcMock, cMock, &auth.NopHandler{})
		api.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Body.String(), ShouldResemble, "incorrect selection of query parameters: [geography], these dimensions do not exist for this version of the dataset\n")

		validateGetDataset(dcMock, "cpih012")
		validateGetVersion(dcMock, "cpih012", "2017", "1")
	})

	Convey("When there is a missing query parameter that is expected to be set in request return 400 bad request with an error message containing a list of missing query parameters", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
		r = r.WithContext(context.WithValue(r.Context(), request.FlorenceIdentityKey, testUserAuthToken))
		w := httptest.NewRecorder()

		dcMock := &mock.IDatasetClientMock{
			GetFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, collectionID string, datasetID string) (dataset.DatasetDetails, error) {
				return dataset.DatasetDetails{State: dataset.StatePublished.String()}, nil
			},
			GetVersionFunc: func(ctx context.Context, userAuthToken string, serviceAuthToken string, downloadServiceAuthToken string, collectionID string, datasetID string, edition string, version string) (dataset.Version, error) {
				return dataset.Version{
					Dimensions: []dataset.VersionDimension{dimension1, dimension2, dimension3, dimension4},
					State:      dataset.StatePublished.String(),
				}, nil
			},
		}

		cMock := &mock.CantabularClientMock{}

		cfg, err := config.Get()
		So(err, ShouldBeNil)
		api := GetAPIWithMocks(cfg, &mock.IGraphMock{}, dcMock, cMock, &auth.NopHandler{})
		api.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Body.String(), ShouldResemble, "missing query parameters for the following dimensions: [age]\n")

		validateGetDataset(dcMock, "cpih012")
		validateGetVersion(dcMock, "cpih012", "2017", "1")
	})

	Convey("When there are too many query parameters that are set to wildcard (*) value request returns 400 bad request", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=*&aggregate=*&geography=K02000001", nil)
		r = r.WithContext(context.WithValue(r.Context(), request.FlorenceIdentityKey, testUserAuthToken))
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

		cMock := &mock.CantabularClientMock{}

		cfg, err := config.Get()
		So(err, ShouldBeNil)
		api := GetAPIWithMocks(cfg, &mock.IGraphMock{}, dcMock, cMock, &auth.NopHandler{})
		api.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Body.String(), ShouldResemble, "only one wildcard (*) is allowed as a value in selected query parameters\n")

		validateGetDataset(dcMock, "cpih012")
		validateGetVersion(dcMock, "cpih012", "2017", "1")
	})

	Convey("When requested query does not find a unique observation return no observations found", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001", nil)
		r = r.WithContext(context.WithValue(r.Context(), request.FlorenceIdentityKey, testUserAuthToken))
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

		cMock := &mock.CantabularClientMock{}

		graphDBMock := &mock.IGraphMock{
			StreamCSVRowsFunc: func(context.Context, string, string, *observation.DimensionFilters, *int) (observation.StreamRowReader, error) {
				return nil, errs.ErrObservationsNotFound
			},
		}

		originalFunc := api.SortFilter
		defer func() {
			api.SortFilter = originalFunc
		}()
		api.SortFilter = func(ctx context.Context, api *api.API, event *models.FilterSubmitted, dbFilter *observation.DimensionFilters) {
		}

		cfg, err := config.Get()
		So(err, ShouldBeNil)
		ap := GetAPIWithMocks(cfg, graphDBMock, dcMock, cMock, &auth.NopHandler{})
		ap.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusNotFound)
		So(w.Body.String(), ShouldContainSubstring, errs.ErrObservationsNotFound.Error())

		validateGetDataset(dcMock, "cpih012")
		validateGetVersion(dcMock, "cpih012", "2017", "1")
		So(len(graphDBMock.StreamCSVRowsCalls()), ShouldEqual, 1)
	})

	Convey("When requested query has a multi-valued dimension return bad request", t, func() {
		r := httptest.NewRequest("GET", "http://localhost:22000/datasets/cpih012/editions/2017/versions/1/observations?time=16-Aug&aggregate=cpi1dim1S40403&geography=K02000001&geography=K02000002", nil)
		r = r.WithContext(context.WithValue(r.Context(), request.FlorenceIdentityKey, testUserAuthToken))
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

		cMock := &mock.CantabularClientMock{}

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		api := GetAPIWithMocks(cfg, &mock.IGraphMock{}, dcMock, cMock, &auth.NopHandler{})
		api.Router.ServeHTTP(w, r)

		So(w.Code, ShouldEqual, http.StatusBadRequest)
		So(w.Body.String(), ShouldResemble, "multi-valued query parameters for the following dimensions: [geography]\n")

		validateGetDataset(dcMock, "cpih012")
		validateGetVersion(dcMock, "cpih012", "2017", "1")
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

func validateGetDataset(dcMock *mock.IDatasetClientMock, datasetID string) {
	So(len(dcMock.GetCalls()), ShouldEqual, 1)
	So(dcMock.GetCalls()[0].ServiceAuthToken, ShouldEqual, testServiceAuthToken)
	So(dcMock.GetCalls()[0].UserAuthToken, ShouldEqual, testUserAuthToken)
	So(dcMock.GetCalls()[0].DatasetID, ShouldEqual, datasetID)
}

func validateGetVersion(dcMock *mock.IDatasetClientMock, datasetID, edition, version string) {
	So(len(dcMock.GetVersionCalls()), ShouldEqual, 1)
	So(dcMock.GetVersionCalls()[0].DatasetID, ShouldEqual, datasetID)
	So(dcMock.GetVersionCalls()[0].Edition, ShouldEqual, edition)
	So(dcMock.GetVersionCalls()[0].Version, ShouldEqual, version)
	So(dcMock.GetVersionCalls()[0].ServiceAuthToken, ShouldEqual, testServiceAuthToken)
	So(dcMock.GetVersionCalls()[0].UserAuthToken, ShouldEqual, testUserAuthToken)
}

func TestSortFilter(t *testing.T) {
	eventFilterSubmitted := models.FilterSubmitted{
		FilterID:   "whatever",
		InstanceID: "460b5039-bb09-4038-b8eb-9091713f4497",
		DatasetID:  "older-people-economic-activity",
		Edition:    "time-series",
		Version:    "1",
	}

	mockRowReader := &observationtest.StreamRowReaderMock{
		ReadFunc: func() (string, error) {
			return "146.3,p,2,Month,Aug-16,K02000001,,cpi1dim1G10100,01.1 Food", nil
		},
		CloseFunc: func(context.Context) error {
			return nil
		},
	}

	graphDBMock := &mock.IGraphMock{
		StreamCSVRowsFunc: func(ctx context.Context, instanceID string, filterID string, filters *observation.DimensionFilters, limit *int) (observation.StreamRowReader, error) {
			return mockRowReader, nil
		},
	}

	var pub = false

	// The following test is to code cover the first return in SortFilter
	Convey("Given a dimension of one, with a mock GetOptions", t, func() {
		var dbFilter = observation.DimensionFilters{
			Dimensions: []*observation.Dimension{
				{
					Name:    "economicactivity",
					Options: []string{"economic-activity", "employment-rate"},
				},
			},
			Published: &pub,
		}

		dcMock := &mock.IDatasetClientMock{
			GetOptionsFunc: func(context.Context, string, string, string, string, string, string, string, *dataset.QueryParams) (dataset.Options, error) {
				return dataset.Options{}, nil
			},
		}

		cMock := &mock.CantabularClientMock{}

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		cfg.ObservationAPIURL = "http://localhost:8082"
		ap := GetAPIWithMocks(cfg, graphDBMock, dcMock, cMock, &auth.NopHandler{})

		Convey("When SortFilter is called", func() {
			api.SortFilter(ctx, ap, &eventFilterSubmitted, &dbFilter)

			Convey("The dimension sees no change", func() {
				So(len(dcMock.GetOptionsCalls()), ShouldEqual, 0)
				So(dbFilter.Dimensions[0].Name, ShouldEqual, "economicactivity")
			})
		})
	})

	Convey("Given a dimension of three, with a mock GetOptions that returns nil simulating error getting record from mongo", t, func() {
		var dbFilter = observation.DimensionFilters{
			Dimensions: []*observation.Dimension{
				{
					Name:    "economicactivity",
					Options: []string{"economic-activity", "employment-rate"},
				},
				{
					Name:    "geography",
					Options: []string{"W92000004"},
				},
				{
					Name:    "sex",
					Options: []string{"people", "men"},
				},
			},
			Published: &pub,
		}

		dcMock := &mock.IDatasetClientMock{
			GetOptionsFunc: func(context.Context, string, string, string, string, string, string, string, *dataset.QueryParams) (dataset.Options, error) {
				return dataset.Options{}, errors.New("can't find record")
			},
		}

		cMock := &mock.CantabularClientMock{}

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		cfg.ObservationAPIURL = "http://localhost:8082"
		ap := GetAPIWithMocks(cfg, graphDBMock, dcMock, cMock, &auth.NopHandler{})

		Convey("When SortFilter is called", func() {
			api.SortFilter(ctx, ap, &eventFilterSubmitted, &dbFilter)

			Convey("The dimension order puts 'geogrphy' first and the rest retain their order", func() {
				// NOTE: As we are simulating mongo errors, depending on how fast the loop in SortFilter
				// manages to run all 3 go routines for the 3 dimensions, sometimes the 1st go routine
				// launched may return the expected error from simulated mongo and exit the loop before
				// the 3rd go routine runs ...
				// which means we can not check the value of GetOptionCalls() as elsewhere as it might
				// return 2 or it might return 3
				// So(len(datasetAPIMock.GetOptionsCalls()), ShouldEqual, 3)
				So(dbFilter.Dimensions[0].Name, ShouldEqual, "geography")
				So(dbFilter.Dimensions[1].Name, ShouldEqual, "economicactivity")
				So(dbFilter.Dimensions[2].Name, ShouldEqual, "sex")
			})
		})
	})

	Convey("Given a dimension of three, with a mock GetOptions that returns size of a Dimension", t, func() {
		var dbFilter = observation.DimensionFilters{
			Dimensions: []*observation.Dimension{
				{
					Name:    "economicactivity",
					Options: []string{"economic-activity", "employment-rate"},
				},
				{
					Name:    "geography",
					Options: []string{"W92000004"},
				},
				{
					Name:    "sex",
					Options: []string{"people", "men"},
				},
			},
			Published: &pub,
		}

		dcMock := &mock.IDatasetClientMock{
			GetOptionsFunc: func(ctx context.Context, userAuthToken, serviceAuthToken, collectionID, id, edition, version, dimension string, q *dataset.QueryParams) (dataset.Options, error) {
				switch dimension {
				case "economicactivity":
					return dataset.Options{TotalCount: 2}, nil // smallest
				case "geography":
					return dataset.Options{TotalCount: 383}, nil // largest
				case "sex":
					return dataset.Options{TotalCount: 3}, nil // in the middle
				}
				return dataset.Options{}, errors.New("can't find record")
			},
		}

		cMock := &mock.CantabularClientMock{}

		cfg, err := config.Get()
		So(err, ShouldBeNil)

		cfg.ObservationAPIURL = "http://localhost:8082"
		ap := GetAPIWithMocks(cfg, graphDBMock, dcMock, cMock, &auth.NopHandler{})

		Convey("When SortFilter is called", func() {
			api.SortFilter(ctx, ap, &eventFilterSubmitted, &dbFilter)

			Convey("The dimension order is returned by largest dimension first to smallest last order", func() {
				So(len(dcMock.GetOptionsCalls()), ShouldEqual, 3)
				So(dbFilter.Dimensions[0].Name, ShouldEqual, "geography")        // largest first
				So(dbFilter.Dimensions[1].Name, ShouldEqual, "sex")              // in the middle
				So(dbFilter.Dimensions[2].Name, ShouldEqual, "economicactivity") // smallest last
			})
		})
	})
}
