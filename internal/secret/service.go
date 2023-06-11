package secret

import (
	"context"
	"errors"
	"time"

	"github.com/fernet/fernet-go"
	"github.com/garaekz/priv8/internal/entity"
	"github.com/garaekz/priv8/pkg/encrypt"
	"github.com/garaekz/priv8/pkg/log"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// MinTTL is the minimum TTL allowed for a secret.
	MinTTL = 5 * time.Minute
)

// Service encapsulates usecase logic for secrets.
type Service interface {
	Get(ctx context.Context, id string) (Secret, error)
	ReadAndBurn(ctx context.Context, id string, req ReadSecretRequest) (DecodedSecret, error)
	Create(ctx context.Context, input CreateSecretRequest) (CreateSecretResponse, error)
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
	Secret     string `json:"secret"`
	Passphrase string `json:"passphrase"`
	TTL        int    `json:"ttl"`
}

// CreateSecretResponse represents an secret creation response.
type CreateSecretResponse struct {
	Code      string `json:"code"`
	Secret    string `json:"secret"`
	ExpiresAt string `json:"expires_at"`
}

// Validate validates the CreateSecretRequest fields.
func (m CreateSecretRequest) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Secret, validation.Required, validation.Length(0, 128)),
		validation.Field(&m.TTL, validation.Required, validation.By(validateTTL)),
	)
}

func validateTTL(value interface{}) error {
	ttl := value.(int)
	if ttl < int(MinTTL.Seconds()) {
		return validation.NewError("ttl", "TTL must be greater than 5 minutes")
	}
	return nil
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
		return DecodedSecret{Code: 404, Error: "Invalid token or token has expired"}, errors.New("Invalid token or token has expired")
	}

	// Burn the secret
	if err = s.repo.Delete(ctx, id); err != nil {
		return DecodedSecret{}, err
	}

	return DecodedSecret{Message: string(message)}, nil
}

// Create creates a new secret.
func (s service) Create(ctx context.Context, req CreateSecretRequest) (CreateSecretResponse, error) {
	if err := req.Validate(); err != nil {
		return CreateSecretResponse{}, err
	}
	id := entity.GenerateID()
	now := time.Now()
	key := encrypt.EncodeKey(req.Passphrase, s.salt)
	token, err := fernet.EncryptAndSign([]byte(req.Secret), &key)
	if err != nil {
		return CreateSecretResponse{}, err
	}

	err = s.repo.Create(ctx, entity.Secret{
		ID:            id,
		EncryptedData: string(token),
		TTL:           req.TTL,
		CreatedAt:     now,
		UpdatedAt:     now,
	})

	if err != nil {
		return CreateSecretResponse{}, err
	}

	return CreateSecretResponse{
		Code:      id,
		Secret:    req.Secret,
		ExpiresAt: now.Add(time.Duration(req.TTL) * time.Second).Format(time.RFC3339),
	}, nil
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
