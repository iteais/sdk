package migrations

import (
	"context"
	"fmt"
	"github.com/iteais/sdk/example/models"
	"github.com/uptrace/bun"
)

func init() {
	Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [up migration] ")

		tx, err := db.BeginTx(ctx, nil)

		if err != nil {
			return err
		}

		_, err = tx.Exec("CREATE SCHEMA IF NOT EXISTS users")

		if err != nil {
			tx.Rollback()
			return err
		}

		user := new(models.User)

		q := tx.NewCreateTable().
			Model(user).
			Table("users.user").
			IfNotExists()

		fmt.Println(q)

		_, err = q.Exec(ctx)

		return tx.Commit()
	}, func(ctx context.Context, db *bun.DB) error {
		fmt.Print(" [down migration] ")
		return nil
	})
}
