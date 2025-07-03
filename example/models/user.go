package models

import (
	"github.com/uptrace/bun"
	"time"
)

type User struct {
	bun.BaseModel `bun:"table:users.user,alias:u"`

	ID        int64  `bun:"id,pk,autoincrement" json:"id"`
	Email     string `bun:"email,notnull" json:"email"`
	FirstName string `bun:"first_name,nullzero" json:"first_name"`

	CreatedAt time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"-"`
	UpdatedAt time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"-"`
	DeletedAt time.Time `bun:"deleted_at,soft_delete,nullzero" json:"-"`
}
