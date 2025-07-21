package pkg

import (
	"fmt"
	"github.com/getsentry/sentry-go"
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
	"strings"
	"sync/atomic"
)

const (
	HealthEndpoint  = "/health"
	MetricsEndpoint = "/metrics"
	ReadyEndpoint   = "/ready"
)

var isReady = &atomic.Value{}

var App *Application

func init() {
	isReady.Store(false)

	if os.Getenv("SENTRY_SERVER") != "" {
		_ = sentry.Init(sentry.ClientOptions{
			Dsn: os.Getenv("SENTRY_SERVER"),
		})
	}
}

type Application struct {
	Db     *bun.DB
	Router *gin.Engine
	Log    *log.Logger
}

type ApplicationConfig struct {
	MigrationPath string
	DbSchemaName  string
	WhiteList     []string
}

func NewApplication(config ApplicationConfig) *Application {

	logger := log.New()
	if strings.ToUpper(os.Getenv("ENVIRONMENT")) != "DEV" {
		logger.Formatter = &log.JSONFormatter{}
	}
	log.SetOutput(logger.Writer())

	dbConn := initDb()
	dbMigrate(config.MigrationPath, config.DbSchemaName)
	dbConn.AddQueryHook(logrusbun.NewQueryHook(logrusbun.QueryHookOptions{Logger: logger}))

	App = &Application{
		Db:     dbConn,
		Router: initRouter(logger),
		Log:    logger,
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

func (a *Application) AppendGetEndpoint(route string, handlers ...gin.HandlerFunc) *Application {
	a.Router.GET(route, handlers...)
	return a
}

func (a *Application) AppendPostEndpoint(route string, handlers ...gin.HandlerFunc) *Application {
	a.Router.POST(route, handlers...)
	return a
}

func (a *Application) AppendPutEndpoint(route string, handlers ...gin.HandlerFunc) *Application {
	a.Router.PUT(route, handlers...)
	return a
}

func (a *Application) AppendDeleteEndpoint(route string, handlers ...gin.HandlerFunc) *Application {
	a.Router.DELETE(route, handlers...)
	return a
}

func (a *Application) AppendPatchEndpoint(route string, handlers ...gin.HandlerFunc) *Application {
	a.Router.PATCH(route, handlers...)
	return a
}

func (a *Application) AppendHeadEndpoint(route string, handlers ...gin.HandlerFunc) *Application {
	a.Router.HEAD(route, handlers...)
	return a
}

func (a *Application) AppendOptionsEndpoint(route string, handlers ...gin.HandlerFunc) *Application {
	a.Router.OPTIONS(route, handlers...)
	return a
}

func (a *Application) AppendSwagger(prefix string) *Application {
	a.Router.GET(prefix+"/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return a
}

func (a *Application) AppendReadyProbe() *Application {
	a.AppendGetEndpoint(ReadyEndpoint, gin.WrapF(ReadyProbe(isReady)))
	return a
}

func (a *Application) AppendMetrics() *Application {
	a.Router.GET(MetricsEndpoint, gin.WrapH(promhttp.Handler()))
	return a
}

func (a *Application) AppendHealthProbe() *Application {
	a.AppendGetEndpoint(HealthEndpoint, HealthProbe(a.Db))
	return a
}

func (a *Application) GetRequestLogger(c *gin.Context) *log.Entry {
	return a.Log.WithField(TraceIdContextKey, c.GetString(TraceIdContextKey))
}

func initRouter(logger *log.Logger) *gin.Engine {
	r := gin.Default()
	r.Use(TraceMiddleware()).
		Use(HttpLogger(logger), gin.Recovery()).
		Use(JsonMiddleware()).
		Use(CorsMiddleware()).
		Use(UserMiddleware())

	return r
}
