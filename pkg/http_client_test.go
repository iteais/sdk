//go:build !exclude_from_test

// exclude because it makes http requests which cant be mocked in ci
package pkg

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFetchEventById(t *testing.T) {
	t.Run("FetchEventById", func(t *testing.T) {
		err := os.Setenv("EVENT_SERVER", "http://localhost:8804")
		if err != nil {
			t.Fatal(err)
		}
		got, err := FetchEventById(1, "traceId", "jwt")
		assert.NoError(t, err)
		assert.Equalf(t, int64(1), got.ID, "FetchEventById(%v, %v, %v)", 1, "traceId", "jwt")
	})
}
