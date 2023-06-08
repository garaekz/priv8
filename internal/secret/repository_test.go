package secret

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/garaekz/priv8/internal/entity"
	"github.com/garaekz/priv8/internal/test"
	"github.com/garaekz/priv8/pkg/log"
	"github.com/stretchr/testify/assert"
)

func TestRepository(t *testing.T) {
	logger, _ := log.NewForTest()
	db := test.DB(t)
	test.ResetTables(t, db, "secret")
	repo := NewRepository(db, logger)

	ctx := context.Background()

	// initial count
	count, err := repo.Count(ctx)
	assert.Nil(t, err)

	// create
	err = repo.Create(ctx, entity.Secret{
		ID:            "test1",
		Identifier:    "test1",
		EncryptedData: "secret1",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	})
	assert.Nil(t, err)
	count2, _ := repo.Count(ctx)
	assert.Equal(t, 1, count2-count)

	// get
	_, err = repo.Get(ctx, "test1")
	assert.Nil(t, err)
	_, err = repo.Get(ctx, "test0")
	assert.Equal(t, sql.ErrNoRows, err)

	// delete
	err = repo.Delete(ctx, "test1")
	assert.Nil(t, err)
	_, err = repo.Get(ctx, "test1")
	assert.Equal(t, sql.ErrNoRows, err)
	err = repo.Delete(ctx, "test1")
	assert.Equal(t, sql.ErrNoRows, err)
}
