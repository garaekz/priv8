package secret

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/garaekz/priv8/internal/entity"
	"github.com/garaekz/priv8/pkg/log"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stretchr/testify/assert"
)

var errCRUD = errors.New("error crud")
var errTTL = validation.NewError("ttl", "TTL must be greater than 5 minutes")

func TestCreateSecretRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		model     CreateSecretRequest
		wantError bool
	}{
		{"success", CreateSecretRequest{Secret: "test", TTL: 300}, false},
		{"required", CreateSecretRequest{Secret: ""}, true},
		{"too long", CreateSecretRequest{Secret: "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"}, true},
		{"ttl < 5min", CreateSecretRequest{Secret: "test", TTL: 30}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.model.Validate()
			assert.Equal(t, tt.wantError, err != nil)
		})
	}
}

func Test_service_CRUD(t *testing.T) {
	logger, _ := log.NewForTest()
	salt := "test"
	s := NewService(&mockRepository{}, logger, salt)

	ctx := context.Background()

	// initial count
	count, _ := s.Count(ctx)
	assert.Equal(t, 0, count)

	// successful creation
	now := time.Now()
	secret, err := s.Create(ctx, CreateSecretRequest{Secret: "test", TTL: 300})
	assert.Nil(t, err)
	assert.NotEmpty(t, secret.Code)
	id := secret.Code

	assert.NotEmpty(t, secret.Secret)
	assert.NotEmpty(t, secret.ExpiresAt)
	assert.Equal(t, "test", secret.Secret)
	assert.Equal(t, now.Add(300*time.Second).Format(time.RFC3339), secret.ExpiresAt)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)

	// validation error in creation
	_, err = s.Create(ctx, CreateSecretRequest{Secret: "", TTL: 0})
	assert.NotNil(t, err)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)

	_, err = s.Create(ctx, CreateSecretRequest{Secret: "error", TTL: 9876543210})
	assert.Equal(t, errCRUD, err)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)

	_, err = s.Create(ctx, CreateSecretRequest{Secret: "error", TTL: 10})
	if vErr, ok := err.(validation.Errors); ok {
		assert.Equal(t, errTTL, vErr["ttl"])
	} else {
		t.Fail()
	}
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)

	_, _ = s.Create(ctx, CreateSecretRequest{Secret: "test2", TTL: 5000})
	count, _ = s.Count(ctx)
	assert.Equal(t, 2, count)

	// get
	_, err = s.Get(ctx, "none")
	assert.NotNil(t, err)
	existingSecret, err := s.Get(ctx, id)
	assert.Nil(t, err)
	assert.Equal(t, id, existingSecret.ID)

	// delete
	_, err = s.Delete(ctx, "none")
	assert.NotNil(t, err)
	existingSecret, err = s.Delete(ctx, id)
	assert.Nil(t, err)
	assert.Equal(t, id, existingSecret.ID)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)
}

func Test_service_ReadAndBurn(t *testing.T) {
	logger, _ := log.NewForTest()
	salt := "test"
	s := NewService(&mockRepository{}, logger, salt)

	ctx := context.Background()

	t.Run("secret doesn't exist or was already read", func(t *testing.T) {
		res, err := s.ReadAndBurn(ctx, "nonexistent", ReadSecretRequest{Passphrase: "pass"})
		assert.NotNil(t, err)
		assert.Equal(t, 404, res.Code)
		assert.Equal(t, "Secret doesn't exist or was already read", res.Error)
	})

	t.Run("invalid token or token has expired", func(t *testing.T) {
		// Create a secret with a very short TTL
		secret, err := s.Create(ctx, CreateSecretRequest{Secret: "test", TTL: 300})
		assert.Nil(t, err)
		if vErr, ok := err.(validation.Errors); ok {
			assert.Equal(t, errTTL, vErr["ttl"])
		} else {
			t.Fail()
		}

		res, err := s.ReadAndBurn(ctx, secret.Code, ReadSecretRequest{Passphrase: "pass"})
		assert.NotNil(t, err)
		assert.Equal(t, 404, res.Code)
		assert.Equal(t, "Invalid token or token has expired", res.Error)
	})

	t.Run("success", func(t *testing.T) {
		// Create a secret
		secret, err := s.Create(ctx, CreateSecretRequest{Secret: "test message", TTL: 300, Passphrase: "test"})
		assert.Nil(t, err)

		res, err := s.ReadAndBurn(ctx, secret.Code, ReadSecretRequest{Passphrase: "test"})
		assert.Nil(t, err)
		assert.Equal(t, "test message", res.Message)
	})

	t.Run("error deleting", func(t *testing.T) {
		res, err := s.ReadAndBurn(ctx, "test", ReadSecretRequest{Passphrase: "pass"})
		assert.NotNil(t, err)
		assert.Empty(t, res.Message)
	})

	t.Run("error burning after reading", func(t *testing.T) {
		// Create a secret
		secret, err := s.Create(ctx, CreateSecretRequest{Secret: "test message", TTL: 999, Passphrase: "test"})
		assert.Nil(t, err)
		assert.NotEmpty(t, secret.Code)

		_, err = s.ReadAndBurn(ctx, "error", ReadSecretRequest{Passphrase: "test"})
		assert.NotNil(t, err)
		assert.Equal(t, errCRUD, err)
	})
}

type mockRepository struct {
	items []entity.Secret
}

func (m *mockRepository) Get(_ context.Context, id string) (entity.Secret, error) {
	for _, item := range m.items {
		if item.ID == id {
			return item, nil
		}
	}
	return entity.Secret{}, sql.ErrNoRows
}

func (m *mockRepository) Count(_ context.Context) (int, error) {
	return len(m.items), nil
}

func (m *mockRepository) Create(_ context.Context, secret entity.Secret) error {
	if secret.TTL == 9876543210 {
		return errCRUD
	}
	if secret.TTL == 999 {
		secret.ID = "error"
	}
	m.items = append(m.items, secret)
	return nil
}

func (m *mockRepository) Delete(_ context.Context, id string) error {
	if id == "error" {
		return errCRUD
	}

	for i, item := range m.items {
		if item.ID == id {
			m.items[i] = m.items[len(m.items)-1]
			m.items = m.items[:len(m.items)-1]
			break
		}
	}
	return nil
}
