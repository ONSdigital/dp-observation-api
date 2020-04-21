package api

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/ONSdigital/dp-observation-api/store"
	storetest "github.com/ONSdigital/dp-observation-api/store/datastoretest"
	"github.com/gorilla/mux"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSetup(t *testing.T) {
	Convey("Given an API instance", t, func() {
		r := mux.NewRouter()
		ctx := context.Background()
		storeMock := &storetest.StorerMock{}
		api := Setup(ctx, r, store.DataStore{Backend: storeMock})

		Convey("When created the following routes should have been added", func() {
			// Replace the check below with any newly added api endpoints
			So(hasRoute(api.Router, "/hello", "GET"), ShouldBeTrue)
			So(hasRoute(api.Router, "/datasets/{dataset_id}/editions/{edition}/versions/{version}/observations", "GET"), ShouldBeTrue)
		})
	})
}

func TestClose(t *testing.T) {
	Convey("Given an API instance", t, func() {
		r := mux.NewRouter()
		ctx := context.Background()
		storeMock := &storetest.StorerMock{}
		a := Setup(ctx, r, store.DataStore{Backend: storeMock})

		Convey("When the api is closed any dependencies are closed also", func() {
			err := a.Close(ctx)
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
