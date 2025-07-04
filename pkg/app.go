package pkg

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/oiime/logrusbun"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/migrate"
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

func NewApplication(dbSchema string, migrations *migrate.Migrations) *Application {

	Logger := log.New()

	dbConn := initDb(dbSchema, migrations)
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

func initDb(dbSchema string, migrations *migrate.Migrations) *bun.DB {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())

	ctx := context.Background()

	db.Exec("CREATE SCHEMA IF NOT EXISTS " + dbSchema + ";")
	tableName := migrate.WithTableName(dbSchema + ".migration")
	lockTableName := migrate.WithLocksTableName(dbSchema + ".migration_lock")

	m := migrate.NewMigrator(db, migrations, tableName, lockTableName)

	err := m.Init(ctx)
	if err != nil {
		panic(err)
	}
	group, err := m.Migrate(ctx)
	//
	if err != nil {
		panic(err)
	}

	if group.IsZero() {
		fmt.Printf("there are no new migrations to run (database is up to date)\n")
	}

	return db
}
