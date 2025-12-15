package pkg

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/iteais/sdk/pkg/app"
	"github.com/minio/minio-go/v7"
	"github.com/oiime/logrusbun"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"github.com/uptrace/bun"
)

const (
	HealthEndpoint  = "/health"
	MetricsEndpoint = "/metrics"
	ReadyEndpoint   = "/ready"
	SwaggerEndpoint = "/swagger/*"
)

var isReady = &atomic.Value{}

var hasSentry = os.Getenv("SENTRY_SERVER") != ""

var App *Application

func init() {
	isReady.Store(false)

	if hasSentry {

		sampleRate := 1.0
		if os.Getenv("ENVIRONMENT") == "DEV" {
			sampleRate = 0.5
		}

		_ = sentry.Init(sentry.ClientOptions{
			Dsn:              os.Getenv("SENTRY_SERVER"),
			Debug:            os.Getenv("ENVIRONMENT") == "DEV",
			Environment:      os.Getenv("ENVIRONMENT"),
			EnableTracing:    true,
			TracesSampleRate: sampleRate,
		})
	}
}

type Application struct {
	Db      *bun.DB
	Router  *gin.Engine
	Log     *log.Logger
	Storage *minio.Client
}

type ApplicationConfig struct {
	AppName       string
	MigrationPath string
	DbSchemaName  string
	WhiteList     []string
}

func NewApplication(config ApplicationConfig) *Application {

	if hasSentry {
		sentry.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetExtra("application", config.AppName)
		})
	}

	logger := log.New()
	if strings.ToUpper(os.Getenv("ENVIRONMENT")) != "DEV" {
		logger.Formatter = &log.JSONFormatter{}
	}
	log.SetOutput(logger.Writer())

	dbConn := app.InitDb()
	app.DbMigrate(config.MigrationPath, config.DbSchemaName)
	dbConn.AddQueryHook(logrusbun.NewQueryHook(logrusbun.QueryHookOptions{Logger: logger}))

	if config.WhiteList == nil {
		config.WhiteList = []string{}
	}

	App = &Application{
		Db:      dbConn,
		Router:  initRouter(logger, config.WhiteList...),
		Log:     logger,
		Storage: app.InitStorage(),
	}

	return App
}

func (a *Application) Run() {

	if hasSentry {
		defer sentry.Flush(time.Second * 5)
	}

	fmt.Println("Application is running")

	a.AppendReadyProbe().AppendHealthProbe().AppendMetrics()

	done := make(chan bool)

	srv := &http.Server{
		Addr:    os.Getenv("HTTP_ADDR"),
		Handler: a.Router,
	}

	go func() {
		log.Printf("Сервер запущен на %s\n", os.Getenv("HTTP_ADDR"))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Ошибка сервера при запуске: %v\n", err)
		}
	}()

	isReady.Store(true)
	<-done
	/* Grace full shutdown*/
	quit := make(chan os.Signal, 1)
	isReady.Store(false)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Получен сигнал остановки. Завершение приложения...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Ошибка при остановке сервера: %v\n", err)
	}

	log.Println("Сервер успешно завершен")
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

func initRouter(logger *log.Logger, whiteList ...string) *gin.Engine {
	r := gin.Default()
	r.Use(TraceMiddleware()).
		Use(HttpLogger(logger), gin.Recovery()).
		Use(JsonMiddleware()).
		Use(CorsMiddleware()).
		Use(UserMiddleware()).
		Use(HmacMiddleware(os.Getenv("HMAC_SERVER"), whiteList...))

	return r
}
