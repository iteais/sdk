package pkg

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
	"github.com/uptrace/bun/migrate"
	"github.com/xid/sdk/example/migrations"
	"os"
)

type Application struct {
	Db         *bun.DB
	Migrations *migrate.Migrations
}

func NewApplication() *Application {

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
	db.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithEnabled(true),
		bundebug.FromEnv(),
	))

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

	return &Application{Db: db}
}

func (a Application) Run() {
	fmt.Println("Application is running")
}
