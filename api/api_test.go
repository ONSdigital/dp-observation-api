package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/ONSdigital/dp-net/request"

	"github.com/ONSdigital/dp-authorisation/auth"
	"github.com/ONSdigital/dp-observation-api/api"
	"github.com/ONSdigital/dp-observation-api/api/mock"
	errs "github.com/ONSdigital/dp-observation-api/apierrors"
	"github.com/ONSdigital/dp-observation-api/config"
	"github.com/gorilla/mux"

	. "github.com/smartystreets/goconvey/convey"
)

const testServiceAuthToken = "testServiceAuthToken"
const testUserAuthToken = "testUserAuthToken"

var ctx = context.WithValue(context.Background(), request.FlorenceIdentityKey, testUserAuthToken)

var (
	mu                sync.Mutex
	testContext       = context.Background()
	codeListAPIURL    = &url.URL{Scheme: "http", Host: "localhost:22400"}
	datasetAPIURL     = &url.URL{Scheme: "http", Host: "localhost:22000"}
	observationAPIURL = &url.URL{Scheme: "http", Host: "localhost:24500"}
)

func TestSetup(t *testing.T) {
	Convey("Given a public API instance", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)
		graphDBMock := &mock.IGraphMock{}
		dcMock := &mock.IDatasetClientMock{}
		cMock := &mock.CantabularClientMock{}
		pMock := &auth.NopHandler{}
		enableURLRewriting := false
		api := GetAPIWithMocks(cfg, graphDBMock, dcMock, cMock, pMock, enableURLRewriting)

		Convey("When created the following routes should have been added", func() {
			So(hasRoute(api.Router, "/datasets/{dataset_id}/editions/{edition}/versions/{version}/observations", "GET"), ShouldBeTrue)
		})
	})

	Convey("Given a private API instance", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)
		graphDBMock := &mock.IGraphMock{}
		dcMock := &mock.IDatasetClientMock{}
		cMock := &mock.CantabularClientMock{}
		pMock := &mock.IAuthHandlerMock{
			RequireFunc: func(required auth.Permissions, handler http.HandlerFunc) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					handler.ServeHTTP(w, r)
				}
			},
		}
		enableURLRewriting := false
		api := GetAPIWithMocks(cfg, graphDBMock, dcMock, cMock, pMock, enableURLRewriting)

		Convey("When created the following routes should have been added", func() {
			So(hasRoute(api.Router, "/datasets/{dataset_id}/editions/{edition}/versions/{version}/observations", "GET"), ShouldBeTrue)
		})
	})
}

func TestClose(t *testing.T) {
	Convey("Given an API instance", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)
		graphDBMock := &mock.IGraphMock{}
		dcMock := &mock.IDatasetClientMock{}
		cMock := &mock.CantabularClientMock{}
		pMock := &auth.NopHandler{}
		enableURLRewriting := false
		api := GetAPIWithMocks(cfg, graphDBMock, dcMock, cMock, pMock, enableURLRewriting)

		Convey("When the api is closed any dependencies are closed also", func() {
			err := api.Close(testContext)
			So(err, ShouldBeNil)
			// Check that dependencies are closed here
		})
	})
}

func hasRoute(r *mux.Router, path, method string) bool {
	req := httptest.NewRequest(method, path, http.NoBody)
	match := &mux.RouteMatch{}
	return r.Match(req, match)
}

// GetAPIWithMocks also used in other tests
func GetAPIWithMocks(cfg *config.Config, graphDBMock api.IGraph, dcMock api.IDatasetClient, cMock api.CantabularClient, pMock api.IAuthHandler, enableURLRewriting bool) *api.API {
	mu.Lock()
	defer mu.Unlock()
	cfg.ServiceAuthToken = testServiceAuthToken
	return api.Setup(testContext, mux.NewRouter(), cfg, graphDBMock, dcMock, cMock, pMock, enableURLRewriting, codeListAPIURL, datasetAPIURL, observationAPIURL)
}

func assertInternalServerErr(w *httptest.ResponseRecorder) {
	So(w.Code, ShouldEqual, http.StatusInternalServerError)
	So(strings.TrimSpace(w.Body.String()), ShouldEqual, errs.ErrInternalServer.Error())
}
