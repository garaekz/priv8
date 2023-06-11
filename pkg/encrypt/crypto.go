package encrypt

import (
	"crypto/sha256"

	"github.com/fernet/fernet-go"
	"golang.org/x/crypto/pbkdf2"
)

// EncodeKey encodes a key with a salt using PBKDF2.
func EncodeKey(keyString string, salt string) fernet.Key {
	keyBytes := pbkdf2.Key([]byte(keyString), []byte(salt), 4096, 32, sha256.New)
	var key fernet.Key
	copy(key[:], keyBytes)
	return key
}
