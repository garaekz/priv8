package secret

import (
	"context"
	"time"

	"github.com/garaekz/priv8/internal/entity"
	"github.com/garaekz/priv8/pkg/log"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// Service encapsulates usecase logic for secrets.
type Service interface {
	Get(ctx context.Context, id string) (Secret, error)
	Create(ctx context.Context, input CreateSecretRequest) (Secret, error)
	Delete(ctx context.Context, id string) (Secret, error)
	Count(ctx context.Context) (int, error)
}

// Secret represents the data about an secret.
type Secret struct {
	entity.Secret
}

// CreateAlbumRequest represents an secret creation request.
type CreateSecretRequest struct {
	RawData string `json:"raw_data"`
}

// Validate validates the CreateAlbumRequest fields.
func (m CreateSecretRequest) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.RawData, validation.Required, validation.Length(0, 128)),
	)
}

// UpdateAlbumRequest represents an secret update request.
type UpdateAlbumRequest struct {
	Name string `json:"name"`
}

// Validate validates the CreateAlbumRequest fields.
func (m UpdateAlbumRequest) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Name, validation.Required, validation.Length(0, 128)),
	)
}

type service struct {
	repo   Repository
	logger log.Logger
}

// NewService creates a new secret service.
func NewService(repo Repository, logger log.Logger) Service {
	return service{repo, logger}
}

// Get returns the secret with the specified the secret ID.
func (s service) Get(ctx context.Context, id string) (Secret, error) {
	secret, err := s.repo.Get(ctx, id)
	if err != nil {
		return Secret{}, err
	}
	return Secret{secret}, nil
}

// Create creates a new secret.
func (s service) Create(ctx context.Context, req CreateSecretRequest) (Secret, error) {
	if err := req.Validate(); err != nil {
		return Secret{}, err
	}
	id := entity.GenerateID()
	now := time.Now()
	err := s.repo.Create(ctx, entity.Secret{
		ID:            id,
		EncryptedData: req.RawData,
		CreatedAt:     now,
		UpdatedAt:     now,
	})
	if err != nil {
		return Secret{}, err
	}
	return s.Get(ctx, id)
}

// Delete deletes the secret with the specified ID.
func (s service) Delete(ctx context.Context, id string) (Secret, error) {
	secret, err := s.Get(ctx, id)
	if err != nil {
		return Secret{}, err
	}
	if err = s.repo.Delete(ctx, id); err != nil {
		return Secret{}, err
	}
	return secret, nil
}

// Count returns the number of albums.
func (s service) Count(ctx context.Context) (int, error) {
	return s.repo.Count(ctx)
}
