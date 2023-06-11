package secret

import (
	"net/http"
	"testing"
	"time"

	"github.com/garaekz/priv8/internal/auth"
	"github.com/garaekz/priv8/internal/entity"
	"github.com/garaekz/priv8/internal/test"
	"github.com/garaekz/priv8/pkg/log"
)

func TestAPI(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)
	salt := "test"
	repo := &mockRepository{items: []entity.Secret{
		{ID: "123", EncryptedData: "secret123", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}}
	RegisterHandlers(router.Group(""), NewService(repo, logger, salt), auth.MockAuthHandler, logger)
	header := auth.MockAuthHeader()

	tests := []test.APITestCase{
		{Name: "get 123", Method: "GET", URL: "/secrets/123", Body: "", Header: nil, WantStatus: http.StatusOK, WantResponse: `*secret123*`},
		{Name: "get unknown", Method: "GET", URL: "/secrets/1234", Body: "", Header: nil, WantStatus: http.StatusNotFound, WantResponse: ""},
		{Name: "create ok", Method: "POST", URL: "/secrets", Body: `{"content":"test", passphrase: null}`, Header: header, WantStatus: http.StatusCreated, WantResponse: "*test*"},
		{Name: "create ok count", Method: "GET", URL: "/secrets", Body: "", Header: nil, WantStatus: http.StatusOK, WantResponse: `*"total_count":2*`},
		{Name: "create auth error", Method: "POST", URL: "/secrets", Body: `{"raw_data":"test"}`, Header: nil, WantStatus: http.StatusUnauthorized, WantResponse: ""},
		{Name: "create input error", Method: "POST", URL: "/secrets", Body: `"raw_data":"test"}`, Header: header, WantStatus: http.StatusBadRequest, WantResponse: ""},
		// {Name: "get all", Method: "GET", URL: "/secrets", Body: "", Header: nil, WantStatus: http.StatusOK, WantResponse: `*"total_count":1*`},
		// {Name: "update ok", Method: "PUT", URL: "/secrets/123", Body: `{"name":"secretxyz"}`, Header: header, WantStatus: http.StatusOK, WantResponse: "*secretxyz*"},
		// {Name: "update verify", Method: "GET", URL: "/secrets/123", Body: "", Header: nil, WantStatus: http.StatusOK, WantResponse: `*secretxyz*`},
		// {Name: "update auth error", Method: "PUT", URL: "/secrets/123", Body: `{"name":"secretxyz"}`, Header: nil, WantStatus: http.StatusUnauthorized, WantResponse: ""},
		// {Name: "update input error", Method: "PUT", URL: "/secrets/123", Body: `"name":"secretxyz"}`, Header: header, WantStatus: http.StatusBadRequest, WantResponse: ""},
		// {Name: "delete ok", Method: "DELETE", URL: "/secrets/123", Body: ``, Header: header, WantStatus: http.StatusOK, WantResponse: "*secretxyz*"},
		// {Name: "delete verify", Method: "DELETE", URL: "/secrets/123", Body: ``, Header: header, WantStatus: http.StatusNotFound, WantResponse: ""},
		// {Name: "delete auth error", Method: "DELETE", URL: "/secrets/123", Body: ``, Header: nil, WantStatus: http.StatusUnauthorized, WantResponse: ""},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}
