package secret

import (
	"context"
	"database/sql"
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
	RegisterHandlers(router.Group(""), &mockService{}, auth.MockAuthHandler, logger)
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
		{Name: "delete ok", Method: "DELETE", URL: "/secrets/123", Body: ``, Header: header, WantStatus: http.StatusOK, WantResponse: "*secret123*"},
		{Name: "delete verify", Method: "DELETE", URL: "/secrets/456", Body: ``, Header: header, WantStatus: http.StatusNotFound, WantResponse: ""},
		{Name: "delete auth error", Method: "DELETE", URL: "/secrets/123", Body: ``, Header: nil, WantStatus: http.StatusUnauthorized, WantResponse: ""},
		{Name: "read and burn ok", Method: "POST", URL: "/secrets/456", Body: `{"passphrase": ""}`, Header: header, WantStatus: http.StatusOK, WantResponse: "*correct*"},
		// {Name: "create auth error", Method: "POST", URL: "/secrets", Body: `{"content":"test"}`, Header: nil, WantStatus: http.StatusUnauthorized, WantResponse: ""},
		// {Name: "get all", Method: "GET", URL: "/secrets", Body: "", Header: nil, WantStatus: http.StatusOK, WantResponse: `*"total_count":1*`},
		// {Name: "update ok", Method: "PUT", URL: "/secrets/123", Body: `{"name":"secretxyz"}`, Header: header, WantStatus: http.StatusOK, WantResponse: "*secretxyz*"},
		// {Name: "update verify", Method: "GET", URL: "/secrets/123", Body: "", Header: nil, WantStatus: http.StatusOK, WantResponse: `*secretxyz*`},
		// {Name: "update auth error", Method: "PUT", URL: "/secrets/123", Body: `{"name":"secretxyz"}`, Header: nil, WantStatus: http.StatusUnauthorized, WantResponse: ""},
		// {Name: "update input error", Method: "PUT", URL: "/secrets/123", Body: `"name":"secretxyz"}`, Header: header, WantStatus: http.StatusBadRequest, WantResponse: ""},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}

// Mock the service
type mockService struct{}

func (*mockService) ReadAndBurn(_ context.Context, _ string, _ ReadSecretRequest) (DecodedSecret, error) {
	return DecodedSecret{
		Message: "correct data",
	}, nil
}

func (*mockService) Create(_ context.Context, _ CreateSecretRequest) (CreateSecretResponse, error) {
	return CreateSecretResponse{
		Code:      "123",
		Secret:    "correct data",
		ExpiresAt: time.Now().Add(time.Second * 300).Format(time.RFC3339),
	}, nil
}

func (*mockService) Get(_ context.Context, id string) (Secret, error) {
	if id == "123" {
		return Secret{
			entity.Secret{
				ID:            id,
				TTL:           300,
				EncryptedData: "secret123",
			},
		}, nil
	}

	return Secret{}, sql.ErrNoRows
}

func (*mockService) Count(_ context.Context) (int, error) {
	return 1, nil
}

func (*mockService) Delete(_ context.Context, id string) (Secret, error) {
	if id == "456" {
		return Secret{}, sql.ErrNoRows
	}

	return Secret{
		entity.Secret{
			ID:            id,
			TTL:           300,
			EncryptedData: "secret123",
		},
	}, nil
}
