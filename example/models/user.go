package models

import (
	"github.com/uptrace/bun"
	"time"
)

type User struct {
	bun.BaseModel `bun:"table:users.user,alias:u" swaggerignore:"true"`

	// Идентификатор пользователя
	ID int64 `bun:"id,pk,autoincrement" json:"id" example:"1"`
	// Почта пользователя
	Email     string `bun:"email,notnull" json:"email" example:"user@example.com"`
	FirstName string `bun:"first_name,nullzero" json:"first_name" example:"John Doe"`

	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"-"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"-"`
	DeletedAt time.Time `bun:"deleted_at,soft_delete,nullzero" json:"-"`
}
