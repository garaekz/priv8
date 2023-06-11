package entity

import (
	"time"
)

// Secret represents a secrets table record.
type Secret struct {
	ID            string    `json:"id"`
	TTL           int       `json:"ttl"`
	EncryptedData string    `json:"ecnrypted_data"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Specify the secret table name.
func (Secret) TableName() string {
	return "secrets"
}
