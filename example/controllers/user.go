package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/xid/sdk/example/models"
	"github.com/xid/sdk/pkg"
)

// GetById godoc
// @Summary      Get user by id
// @Description  Get user by id
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "user id"
// @Success      200  {object}  models.User
// @Router       /user/{id} [get]
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

		pkg.App.GetRequestLogger(c).Info(query.String())

		err := query.Scan(c)
		pkg.App.GetRequestLogger(c).Info("some else msg")
		c.JSON(200, gin.H{"data": user, "error": err})
	}
}
