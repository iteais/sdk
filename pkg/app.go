package pkg

import (
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/oiime/logrusbun"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"github.com/uptrace/bun"
	"os"
	"sync/atomic"
)

var isReady = &atomic.Value{}

var App *Application

func init() {
	isReady.Store(false)
}

type Application struct {
	Db     *bun.DB
	Router *gin.Engine
	Log    *log.Logger
}

type ApplicationConfig struct {
	MigrationPath string
	DbSchemaName  string
}

func NewApplication(config ApplicationConfig) *Application {

	Logger := log.New()

	dbConn := initDb()
	dbMigrate(config.MigrationPath, config.DbSchemaName)
	dbConn.AddQueryHook(logrusbun.NewQueryHook(logrusbun.QueryHookOptions{Logger: Logger}))

	App = &Application{
		Db:     dbConn,
		Router: initRouter(Logger),
		Log:    Logger,
	}

	return App
}

func (a *Application) Run() {
	fmt.Println("Application is running")

	a.AppendReadyProbe().AppendHealthProbe().AppendMetrics()

	done := make(chan bool)
	go a.Router.Run(os.Getenv("HTTP_ADDR"))
	isReady.Store(true)
	<-done
}

func (a *Application) AppendGetEndpoint(route string, handler gin.HandlerFunc) *Application {
	a.Router.GET(route, handler)
	return a
}

func (a *Application) AppendPostEndpoint(route string, handler gin.HandlerFunc) *Application {
	a.Router.POST(route, handler)
	return a
}

func (a *Application) AppendPutEndpoint(route string, handler gin.HandlerFunc) *Application {
	a.Router.PUT(route, handler)
	return a
}

func (a *Application) AppendDeleteEndpoint(route string, handler gin.HandlerFunc) *Application {
	a.Router.DELETE(route, handler)
	return a
}

func (a *Application) AppendPatchEndpoint(route string, handler gin.HandlerFunc) *Application {
	a.Router.PATCH(route, handler)
	return a
}

func (a *Application) AppendHeadEndpoint(route string, handler gin.HandlerFunc) *Application {
	a.Router.HEAD(route, handler)
	return a
}

func (a *Application) AppendOptionsEndpoint(route string, handler gin.HandlerFunc) *Application {
	a.Router.OPTIONS(route, handler)
	return a
}

func (a *Application) AppendSwagger(prefix string) *Application {
	a.Router.GET(prefix+"/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return a
}

func (a *Application) AppendReadyProbe() *Application {
	a.AppendGetEndpoint("/ready", gin.WrapF(ReadyProbe(isReady)))
	return a
}

func (a *Application) AppendMetrics() *Application {
	a.Router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	return a
}

func (a *Application) AppendHealthProbe() *Application {
	a.AppendGetEndpoint("/health", HealthProbe(a.Db))
	return a
}

func (a *Application) GetRequestLogger(c *gin.Context) *log.Entry {
	return a.Log.WithField(traceIdContextKey, c.GetString(traceIdContextKey))
}

func initRouter(logger *log.Logger) *gin.Engine {
	r := gin.Default()
	r.Use(TraceMiddleware()).
		Use(HttpLogger(logger), gin.Recovery()).
		Use(JsonMiddleware()).
		Use(CorsMiddleware())

	return r
}
