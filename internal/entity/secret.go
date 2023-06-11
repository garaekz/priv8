package entity

import (
	"time"
)

// Secret represents a secrets table record.
type Secret struct {
	ID            string    `json:"id"`
	TTL           int       `json:"ttl"`
	EncryptedData string    `json:"encrypted_data"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TableName specifies the name of the table.
func (Secret) TableName() string {
	return "secrets"
}
