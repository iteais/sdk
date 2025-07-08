package pkg

import (
	"crypto/sha256"
	"encoding/hex"
)

type ApiAccount struct {
	ID          int64  `json:"id"`
	Key         string `json:"key"`
	Secret      string `json:"secret"`
	Role        string `json:"role"`
	Block       bool   `json:"block"`
	BlockReason string `json:"block_reason"`
	Comment     string `json:"comment"`
}

// CanHandleWithHash verifies if the provided hash matches the expected hash for the given time.
// Returns true if the hash matches, false otherwise.
func (a ApiAccount) CanHandleWithHash(hash string, time string) bool {
	if a.Block {
		return false
	}
	return hash == a.GetHash(time)
}

// GetHash generates a SHA-256 hash based on the account's key, secret, and the provided time.
// The hash is generated from the concatenation of key + time + secret.
func (a ApiAccount) GetHash(time string) string {
	hasher := sha256.New()
	data := []byte(a.Key + time + a.Secret)
	hasher.Write(data)

	return hex.EncodeToString(hasher.Sum(nil))
}
