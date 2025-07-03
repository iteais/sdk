package pkg

import (
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"net/http"
	"sync/atomic"
)

func HealthProbe(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := db.Exec("SELECT 1 = 1")

		if err != nil {
			c.AbortWithStatus(http.StatusServiceUnavailable)
		}

		c.Writer.WriteHeader(http.StatusOK)
	}
}

func ReadyProbe(isReady *atomic.Value) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		if isReady == nil || !isReady.Load().(bool) {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
