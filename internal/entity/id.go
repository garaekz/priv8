package entity

import (
	"math/rand"
	"strings"
	"time"

	ulid "github.com/oklog/ulid/v2"
)

// GenerateID generates a unique ID that can be used as an identifier for an entity.
func GenerateID() string {
	entropy := rand.New(rand.NewSource(time.Now().UnixNano()))
	ms := ulid.Timestamp(time.Now())
	return strings.ToLower(ulid.MustNew(ms, entropy).String())
}
