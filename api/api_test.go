package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	errs "github.com/ONSdigital/dp-observation-api/apierrors"
	"github.com/ONSdigital/dp-observation-api/config"
	"github.com/ONSdigital/dp-observation-api/store"
	storetest "github.com/ONSdigital/dp-observation-api/store/datastoretest"
	"github.com/gorilla/mux"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	mu          sync.Mutex
	testContext = context.Background()
)

func TestSetup(t *testing.T) {
	Convey("Given an API instance", t, func() {
		storeMock := &storetest.StorerMock{}
		api := GetAPIWithMocks(storeMock)

		Convey("When created the following routes should have been added", func() {
			So(hasRoute(api.Router, "/datasets/{dataset_id}/editions/{edition}/versions/{version}/observations", "GET"), ShouldBeTrue)
		})
	})
}

func TestClose(t *testing.T) {
	Convey("Given an API instance", t, func() {
		storeMock := &storetest.StorerMock{}
		api := GetAPIWithMocks(storeMock)

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

// GetAPIWithMocks also used in other tests, so exported
func GetAPIWithMocks(storeMock store.Storer) *API {
	mu.Lock()
	defer mu.Unlock()
	cfg, err := config.Get()
	So(err, ShouldBeNil)
	return Setup(testContext, mux.NewRouter(), cfg, store.DataStore{Backend: storeMock})
}

func assertInternalServerErr(w *httptest.ResponseRecorder) {
	So(w.Code, ShouldEqual, http.StatusInternalServerError)
	So(strings.TrimSpace(w.Body.String()), ShouldEqual, errs.ErrInternalServer.Error())
}
