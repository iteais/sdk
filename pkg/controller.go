package pkg

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"math"
	"net/http"
	"strconv"
	"strings"
)

func ListAction[T interface{}]() func(c *gin.Context) {
	return func(c *gin.Context) {

		modelsArray := make([]T, 0)

		perPageParam := c.DefaultQuery("per-page", "20")
		perPage, _ := strconv.Atoi(perPageParam)

		pageParam := c.DefaultQuery("page", "1")
		page, _ := strconv.Atoi(pageParam)

		query := App.Db.NewSelect().
			Model(&modelsArray).
			Limit(perPage).
			Offset((page - 1) * perPage)

		fields := c.Query("fields")
		if fields != "" {
			query = query.ColumnExpr(fields)
		}

		sort := c.Query("sort")

		if sort != "" {
			direction := "DESC"
			if strings.HasPrefix(sort, "-") {
				direction = "ASC"
			}

			sortField := strings.Replace(sort, "-", "", 1)

			query = query.Order(sortField + " " + direction)
		}

		count, err := query.ScanAndCount(context.Background())

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if count < 1 {
			c.JSON(http.StatusNotFound, gin.H{})
			return
		}

		c.Header("X-Total-Count", fmt.Sprintf("%d", count))
		c.Header("x-pagination-per-page", fmt.Sprintf("%d", perPage))

		xppc := 1
		calcXppc := math.Round(float64(count / perPage))
		if calcXppc > 0 {
			xppc = int(calcXppc)
		}

		c.Header("x-pagination-page-count", fmt.Sprintf("%d", xppc))

		xpcp := 1
		calcXpcp := xppc - page
		if calcXpcp > 0 {
			xpcp = int(calcXpcp)
		}

		//x-pagination-current-page
		c.Header("x-pagination-current-page", fmt.Sprintf("%d", xpcp))

		c.JSON(200, gin.H{"data": modelsArray})
	}
}
