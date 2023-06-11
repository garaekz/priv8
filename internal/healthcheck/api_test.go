package healthcheck

import (
	"net/http"
	"testing"

	"github.com/garaekz/priv8/internal/test"
	"github.com/garaekz/priv8/pkg/log"
)

func TestAPI(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)
	RegisterHandlers(router, "0.9.0")
	test.Endpoint(t, router, test.APITestCase{
		Name: "ok", Method: "GET", URL: "/healthcheck", Body: "", Header: nil, WantStatus: http.StatusOK, WantResponse: `"OK 0.9.0"`,
	})
}
