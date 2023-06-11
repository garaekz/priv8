package secret

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/garaekz/priv8/internal/auth"
	"github.com/garaekz/priv8/internal/entity"
	"github.com/garaekz/priv8/internal/test"
	"github.com/garaekz/priv8/pkg/log"
	"github.com/stretchr/testify/assert"
)

func TestSecretsAPI(t *testing.T) {
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
		{
			Name:       "create ok",
			Method:     "POST",
			URL:        "/secrets",
			Body:       `{"secret":"correct data", "ttl":300}`,
			Header:     header,
			WantStatus: http.StatusCreated,
			AssertFunc: func(t *testing.T, res *http.Response) {
				bodyBytes, _ := ioutil.ReadAll(res.Body)
				var body map[string]interface{}
				_ = json.Unmarshal(bodyBytes, &body)
				assert.NotEmpty(t, body["code"])
				assert.NotEmpty(t, body["secret"])
				assert.Equal(t, "correct data", body["secret"])
				assert.NotEmpty(t, body["expires_at"])
			},
		},
		{Name: "create input error", Method: "POST", URL: "/secrets", Body: `"secret":"test"}`, Header: header, WantStatus: http.StatusBadRequest, WantResponse: ""},
		// {Name: "create auth error", Method: "POST", URL: "/secrets", Body: `{"content":"test"}`, Header: nil, WantStatus: http.StatusUnauthorized, WantResponse: ""},
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
