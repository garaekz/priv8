package secret

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/garaekz/priv8/internal/entity"
	"github.com/garaekz/priv8/pkg/log"
	"github.com/stretchr/testify/assert"
)

var errCRUD = errors.New("error crud")

func TestCreateSecretRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		model     CreateSecretRequest
		wantError bool
	}{
		{"success", CreateSecretRequest{RawData: "test"}, false},
		{"required", CreateSecretRequest{RawData: ""}, true},
		{"too long", CreateSecretRequest{RawData: "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"}, true},
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
	s := NewService(&mockRepository{}, logger)

	ctx := context.Background()

	// initial count
	count, _ := s.Count(ctx)
	assert.Equal(t, 0, count)

	// successful creation
	secret, err := s.Create(ctx, CreateSecretRequest{RawData: "test"})
	assert.Nil(t, err)
	assert.NotEmpty(t, secret.ID)
	id := secret.ID
	assert.NotEmpty(t, secret.CreatedAt)
	assert.NotEmpty(t, secret.UpdatedAt)
	assert.Equal(t, 1, count)

	// validation error in creation
	_, err = s.Create(ctx, CreateSecretRequest{RawData: ""})
	assert.NotNil(t, err)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)

	// unexpected error in creation
	_, err = s.Create(ctx, CreateSecretRequest{RawData: "error"})
	assert.Equal(t, errCRUD, err)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)

	_, _ = s.Create(ctx, CreateSecretRequest{RawData: "test2"})

	// get
	_, err = s.Get(ctx, "none")
	assert.NotNil(t, err)
	secret, err = s.Get(ctx, id)
	assert.Nil(t, err)
	assert.Equal(t, id, secret.ID)

	// delete
	_, err = s.Delete(ctx, "none")
	assert.NotNil(t, err)
	secret, err = s.Delete(ctx, id)
	assert.Nil(t, err)
	assert.Equal(t, id, secret.ID)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)
}

type mockRepository struct {
	items []entity.Secret
}

func (m mockRepository) Get(_ context.Context, id string) (entity.Secret, error) {
	for _, item := range m.items {
		if item.ID == id {
			return item, nil
		}
	}
	return entity.Secret{}, sql.ErrNoRows
}

func (m mockRepository) Count(_ context.Context) (int, error) {
	return len(m.items), nil
}

func (m *mockRepository) Create(_ context.Context, secret entity.Secret) error {
	if secret.EncryptedData == "error" {
		return errCRUD
	}
	m.items = append(m.items, secret)
	return nil
}

func (m *mockRepository) Delete(_ context.Context, id string) error {
	for i, item := range m.items {
		if item.ID == id {
			m.items[i] = m.items[len(m.items)-1]
			m.items = m.items[:len(m.items)-1]
			break
		}
	}
	return nil
}
