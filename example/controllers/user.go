package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"github.com/iteais/sdk/example/models"
	"github.com/iteais/sdk/pkg"
	"github.com/iteais/sdk/pkg/utils"
	"io"
	"strings"
)

// SplitAndTrim splits a string by sep and trims whitespace from each element.
func SplitAndTrim(s, sep string) []string {
	parts := []string{}
	for _, p := range strings.Split(s, sep) {
		p = strings.TrimSpace(p)
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}

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
		// Define allowed columns for user model
		allowedColumns := map[string]bool{
			"id":    true,
			"name":  true,
			"email": true,
			// Add other allowed columns here
		}

		fields := c.Query("fields")

		user := new(models.User)
		query := pkg.App.Db.NewSelect().
			Model(user).
			Where("id = ?", id)

		if fields != "" {
			// Split fields by comma and validate each
			validFields := []string{}
			for _, f := range SplitAndTrim(fields, ",") {
				if allowedColumns[f] {
					validFields = append(validFields, f)
				}
			}
			if len(validFields) > 0 {
				query = query.Column(validFields...)
			}
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
			JWT:     c.GetString(utils.AuthHeader),
		})

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return
		}

		c.Render(resp.StatusCode, render.Data{Data: body})
	}
}
