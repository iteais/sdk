package pkg

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/oiime/logrusbun"
	log "github.com/sirupsen/logrus"
	"github.com/toorop/gin-logrus"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/migrate"
	"github.com/xid/sdk/example/migrations"
	"net/http"
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

func NewApplication() *Application {

	Logger := log.New()

	dbConn := initDb()
	dbConn.AddQueryHook(logrusbun.NewQueryHook(logrusbun.QueryHookOptions{Logger: Logger}))

	App = &Application{
		Db:     dbConn,
		Router: initRouter(Logger),
		Log:    Logger,
	}

	return App
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

func initRouter(logger *log.Logger) *gin.Engine {
	r := gin.Default()
	r.Use(ginlogrus.Logger(logger), gin.Recovery())

	return r
}

func initDb() *bun.DB {
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

	m := migrate.NewMigrator(db, migrations.Migrations)
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
