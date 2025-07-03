package pkg

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/uptrace/bun"
	"net/http"
	"os"
	"sync/atomic"
)

var isReady = &atomic.Value{}

func init() {
	isReady.Store(false)
}

type Application struct {
	Db     *bun.DB
	Router *gin.Engine
}

func NewApplication() *Application {
	return &Application{
		Db:     Db,
		Router: initRouter(),
	}
}

func (a Application) Run() {
	fmt.Println("Application is running")

	a.AppendReadyProbe()
	a.AppendHealthProbe()

	done := make(chan bool)
	go a.Router.Run(os.Getenv("HTTP_ADDR"))
	isReady.Store(true)
	<-done
}

func (a Application) AppendGetEndpoint(route string, handler gin.HandlerFunc) {
	a.Router.GET(route, handler)
}

func (a Application) AppendReadyProbe() {
	a.AppendGetEndpoint("/ready", gin.WrapF(ReadyProbe(isReady)))
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

func (a Application) AppendHealthProbe() {
	a.AppendGetEndpoint("/health", HealthWith(a.Db))
}

func HealthWith(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := db.Exec("SELECT 1 = 1")

		if err != nil {
			c.AbortWithStatus(http.StatusServiceUnavailable)
		}

		c.Writer.WriteHeader(http.StatusOK)
	}
}

func initRouter() *gin.Engine {
	r := gin.Default()

	return r
}
