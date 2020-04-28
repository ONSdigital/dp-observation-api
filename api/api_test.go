package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/ONSdigital/dp-observation-api/api"
	"github.com/ONSdigital/dp-observation-api/api/mock"
	errs "github.com/ONSdigital/dp-observation-api/apierrors"
	"github.com/ONSdigital/dp-observation-api/config"
	"github.com/gorilla/mux"

	. "github.com/smartystreets/goconvey/convey"
)

const testServiceAuthToken = "testServiceAuthToken"

var (
	mu          sync.Mutex
	testContext = context.Background()
)

func TestSetup(t *testing.T) {
	Convey("Given an API instance", t, func() {
		cfg, err := config.Get()
		So(err, ShouldBeNil)
		graphDBMock := &mock.IGraphMock{}
		dcMock := &mock.IDatasetClientMock{}
		api := GetAPIWithMocks(graphDBMock, dcMock, cfg)

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
		api := GetAPIWithMocks(graphDBMock, dcMock, cfg)

		Convey("When the api is closed any dependencies are closed also", func() {
			err := api.Close(testContext)
			So(err, ShouldBeNil)
			// Check that dependencies are closed here
		})
	})
}

func hasRoute(r *mux.Router, path, method string) bool {
	req := httptest.NewRequest(method, path, nil)
	match := &mux.RouteMatch{}
	return r.Match(req, match)
}

// GetAPIWithMocks also used in other tests
func GetAPIWithMocks(graphDBMock api.IGraph, dcMock api.IDatasetClient, cfg *config.Config) *api.API {
	mu.Lock()
	defer mu.Unlock()
	return api.Setup(testContext, mux.NewRouter(), cfg, graphDBMock, dcMock)
}

func assertInternalServerErr(w *httptest.ResponseRecorder) {
	So(w.Code, ShouldEqual, http.StatusInternalServerError)
	So(strings.TrimSpace(w.Body.String()), ShouldEqual, errs.ErrInternalServer.Error())
}
