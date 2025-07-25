package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"github.com/iteais/sdk/example/models"
	"github.com/iteais/sdk/pkg"
	"io"
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

// Proxy Internal request example
func Proxy() gin.HandlerFunc {
	return func(c *gin.Context) {
		pkg.App.GetRequestLogger(c).Info("Proxy")

		resp := pkg.InternalFetch(pkg.InternalFetchConfig{
			Method:  "GET",
			Url:     "http://localhost:8800/user/1",
			TraceId: c.GetString(pkg.TraceIdContextKey),
			JWT:     c.GetString("Authorization"),
		})

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return
		}

		c.Render(resp.StatusCode, render.Data{Data: body})
	}
}
