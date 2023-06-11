package secret

import (
	"context"
	"time"

	"github.com/fernet/fernet-go"
	"github.com/garaekz/priv8/internal/entity"
	"github.com/garaekz/priv8/pkg/encrypt"
	"github.com/garaekz/priv8/pkg/log"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// Service encapsulates usecase logic for secrets.
type Service interface {
	Get(ctx context.Context, id string) (Secret, error)
	ReadAndBurn(ctx context.Context, id string, req ReadSecretRequest) (DecodedSecret, error)
	Create(ctx context.Context, input CreateSecretRequest) (Secret, error)
	Delete(ctx context.Context, id string) (Secret, error)
	Count(ctx context.Context) (int, error)
}

// Secret represents the data about an secret.
type Secret struct {
	entity.Secret
}

// DecodedSecret represents the response of a decoded secret.
type DecodedSecret struct {
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
	Error   string `json:"error,omitempty"`
}

// CreateSecretRequest represents an secret creation request.
type CreateSecretRequest struct {
	Content        string  `json:"content"`
	Passphrase     string  `json:"passphrase"`
	ExpirationTime *string `json:"expiration_time"`
}

// Validate validates the CreateSecretRequest fields.
func (m CreateSecretRequest) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Content, validation.Required, validation.Length(0, 128)),
	)
}

// ReadSecretRequest represents an secret reading request.
type ReadSecretRequest struct {
	Passphrase string `json:"passphrase"`
}

type service struct {
	repo   Repository
	logger log.Logger
	salt   string
}

// NewService creates a new secret service.
func NewService(repo Repository, logger log.Logger, salt string) Service {
	return service{repo, logger, salt}
}

// Get returns the secret with the specified the secret ID.
func (s service) Get(ctx context.Context, id string) (Secret, error) {
	secret, err := s.repo.Get(ctx, id)
	if err != nil {
		return Secret{}, err
	}
	return Secret{secret}, nil
}

func (s service) ReadAndBurn(ctx context.Context, id string, req ReadSecretRequest) (DecodedSecret, error) {
	secret, err := s.repo.Get(ctx, id)
	if err != nil {
		return DecodedSecret{Code: 404, Error: "Secret doesn't exist or was already read"}, nil
	}

	key := encrypt.EncodeKey(req.Passphrase, s.salt)
	ttl := time.Duration(secret.TTL) * time.Second
	message := fernet.VerifyAndDecrypt([]byte(secret.EncryptedData), ttl, []*fernet.Key{&key})
	if message == nil {
		return DecodedSecret{Code: 404, Error: "Invalid token or token has expired"}, nil
	}

	// Burn the secret
	if err = s.repo.Delete(ctx, id); err != nil {
		return DecodedSecret{}, err
	}

	return DecodedSecret{Message: string(message)}, nil
}

// Create creates a new secret.
func (s service) Create(ctx context.Context, req CreateSecretRequest) (Secret, error) {
	if err := req.Validate(); err != nil {
		return Secret{}, err
	}
	id := entity.GenerateID()
	now := time.Now()
	key := encrypt.EncodeKey(req.Passphrase, s.salt)
	token, err := fernet.EncryptAndSign([]byte(req.Content), &key)
	if err != nil {
		return Secret{}, err
	}
	// TTL is now a constant of 60 seconds, this will be changed in the near future
	ttl := 60 * time.Second
	err = s.repo.Create(ctx, entity.Secret{
		ID:            id,
		EncryptedData: string(token),
		TTL:           int(ttl.Seconds()),
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
