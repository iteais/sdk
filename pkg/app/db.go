package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"os"
)

func getDbDsn() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
}

func InitDb() *bun.DB {
	dsn := getDbDsn()

	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqldb, pgdialect.New())
	return db
}

func DbMigrate(path string, schemaName string) {

	dsn := getDbDsn()

	db, err := sql.Open("postgres", dsn)
	conn, err := db.Conn(context.Background())
	driver, err := postgres.WithConnection(context.Background(), conn, &postgres.Config{MigrationsTable: "migrations", SchemaName: schemaName})

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+path,
		"postgres",
		driver,
	)

	if err != nil {
		panic(err)
	}

	err = m.Up() // run your migrations and handle the errors above of course
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		panic(err)
	}

}
