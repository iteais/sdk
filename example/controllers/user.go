package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/xid/sdk/example/models"
	"github.com/xid/sdk/pkg"
)

func GetById() gin.HandlerFunc {
	return func(c *gin.Context) {

		id := c.Param("id")
		fields := c.Query("fields")

		user := new(models.User)
		query := pkg.App.Db.NewSelect().
			Model(user).
			Where("id = ?", id)

		if fields != "" {
			query = query.ColumnExpr(fields)
		}

		pkg.App.Log.Info(query.String())

		err := query.Scan(c)
		c.JSON(200, gin.H{"data": user, "error": err})
	}
}
