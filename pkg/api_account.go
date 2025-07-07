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

func (a ApiAccount) CanHandeWithHash(hash string, time string) bool {
	return hash == a.GetHash(time)
}

func (a ApiAccount) GetHash(time string) string {
	hasher := sha256.New()
	data := []byte(a.Key + time + a.Secret)
	hasher.Write(data)

	// Get the hash sum as a byte slice
	hashBytes := hasher.Sum(nil)

	// Convert the hash bytes to a hexadecimal string
	return hex.EncodeToString(hashBytes)
}
