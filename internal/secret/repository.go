package secret

import (
	"context"

	"github.com/garaekz/priv8/internal/entity"
	"github.com/garaekz/priv8/pkg/dbcontext"
	"github.com/garaekz/priv8/pkg/log"
)

// Repository encapsulates the logic to access secrets from the data source.
type Repository interface {
	// Get returns the secret with the specified secret ID.
	Get(ctx context.Context, id string) (entity.Secret, error)
	// Create saves a new secret in the storage.
	Create(ctx context.Context, secret entity.Secret) error
	// Delete removes the secret with given ID from the storage.
	Delete(ctx context.Context, id string) error
	// Count returns the number of secrets.
	Count(ctx context.Context) (int, error)
}

// repository persists secrets in database
type repository struct {
	db     *dbcontext.DB
	logger log.Logger
}

// NewRepository creates a new secret repository
func NewRepository(db *dbcontext.DB, logger log.Logger) Repository {
	return repository{db, logger}
}

// Get reads the secret with the specified ID from the database.
func (r repository) Get(ctx context.Context, id string) (entity.Secret, error) {
	var secret entity.Secret
	err := r.db.With(ctx).Select().Model(id, &secret)
	return secret, err
}

// Create saves a new secret record in the database.
// It returns the ID of the newly inserted secret record.
func (r repository) Create(ctx context.Context, secret entity.Secret) error {
	return r.db.With(ctx).Model(&secret).Insert()
}

// Delete deletes an secret with the specified ID from the database.
func (r repository) Delete(ctx context.Context, id string) error {
	secret, err := r.Get(ctx, id)
	if err != nil {
		return err
	}
	return r.db.With(ctx).Model(&secret).Delete()
}

// Count returns the number of the album records in the database.
func (r repository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.With(ctx).Select("COUNT(*)").From("secrets").Row(&count)
	return count, err
}
