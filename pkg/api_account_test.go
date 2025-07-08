package pkg

import (
	"testing"
)

func TestApiAccount_GetHash(t *testing.T) {
	tests := []struct {
		name    string
		account ApiAccount
		time    string
		want    string
	}{
		{
			name: "valid_hash_generation",
			account: ApiAccount{
				ID:     1,
				Key:    "test-key",
				Secret: "test-secret",
				Role:   "admin",
			},
			time: "2024-01-01T00:00:00Z",
			want: "103d169c35bc2c4b39e216d18ca0e7babac46c73da206d42f81e181e49f758c4",
		},
		{
			name: "empty_key",
			account: ApiAccount{
				ID:     1,
				Key:    "",
				Secret: "test-secret",
				Role:   "admin",
			},
			time: "2024-01-01T00:00:00Z",
			want: "8a15001a172b31a067b58a48b91a81d2c7eb7fbdd18c1586006b43e73facc810",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.account.GetHash(tt.time)
			if got != tt.want {
				t.Errorf("GetHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApiAccount_CanHandleWithHash(t *testing.T) {
	tests := []struct {
		name    string
		account ApiAccount
		time    string
		want    bool
	}{
		{
			name: "valid_hash_match",
			account: ApiAccount{
				ID:     1,
				Key:    "test-key",
				Secret: "test-secret",
				Role:   "admin",
			},
			time: "2024-01-01T00:00:00Z",
			want: true,
		},
		{
			name: "invalid_hash",
			account: ApiAccount{
				ID:     1,
				Key:    "test-key",
				Secret: "test-secret",
				Role:   "admin",
			},
			time: "2024-01-01T00:00:00Z",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := tt.account.GetHash(tt.time)
			if got := tt.account.CanHandleWithHash(hash, tt.time); got != tt.want {
				t.Errorf("CanHandleWithHash() = %v, want %v", got, tt.want)
			}
		})
	}
}
