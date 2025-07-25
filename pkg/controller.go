package pkg

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/iteais/sdk/pkg/models"
	"github.com/iteais/sdk/pkg/utils"
	"github.com/uptrace/bun"
	"math"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

func ListAction[T struct{}](postFindFuncs ...func(*gin.Context, *[]T)) func(c *gin.Context) {
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

		expand := c.Query("expand")
		if expand != "" {

			relations := strings.Split(expand, ",")

			for _, relation := range relations {
				query = query.Relation(strings.TrimSpace(relation))
			}
		}

		ApplyFilter[T](c, query)

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			appendLastModifiedHeader[T](c, query)
		}()

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

		for _, f := range postFindFuncs {
			f(c, &modelsArray)
		}

		//x-pagination-current-page
		c.Header("x-pagination-current-page", fmt.Sprintf("%d", xpcp))
		wg.Wait()
		c.JSON(200, gin.H{"data": modelsArray})
	}
}

func appendLastModifiedHeader[T struct{}](c *gin.Context, query *bun.SelectQuery) {
	t := new(T)
	typ := reflect.TypeOf(t)

	field := ""

	if _, found := typ.FieldByName("UpdatedAt"); found {
		field = "updated_at"
	} else if _, found := typ.FieldByName("CreatedAt"); found {
		field = "created_at"
	}

	if field != "" {

		var timeString string

		query.ColumnExpr(field).Order("? DESC", field).Scan(c, timeString)
		layout := "2006-01-02 15:04:05"

		dt, err := time.Parse(layout, timeString)

		if err != nil {
			c.Header("Last-Modified", dt.Format(time.RFC1123))
		}
	}
}

func ApplyFilter[T interface{}](c *gin.Context, query *bun.SelectQuery) {

	filter, exists := c.GetQueryMap("filter")
	if exists == false {
		return
	}

	model := new(T)
	structValue := reflect.ValueOf(model)

	for key, value := range filter {
		if key == "" || value == "" {
			continue
		}

		methodName := "By" + utils.ToUpperCamelCase(key)
		method := structValue.MethodByName(methodName)

		if method.IsValid() != false {
			args := []reflect.Value{reflect.ValueOf(value), reflect.ValueOf(query)}
			method.Call(args)
		}

		additionalFilter := structValue.MethodByName("CommonListFilter")

		if additionalFilter.IsValid() != false {
			args := []reflect.Value{reflect.ValueOf(c), reflect.ValueOf(query)}
			additionalFilter.Call(args)
		}
	}

}

func UpdateAction[T interface{}](pk string) func(*gin.Context) {
	return func(c *gin.Context) {
		id := c.Param("id")

		existModel := new(T)
		query := App.Db.NewSelect().
			Model(existModel).
			Where("? = ?", bun.Ident(pk), id)
		count, err := query.ScanAndCount(c)

		if count < 1 {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		newModel, loadErrors := models.LoadModel(c, existModel, make(map[string]string))

		if len(loadErrors) > 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"errors": loadErrors})
			return
		}

		q := App.Db.NewUpdate().
			Model(newModel).
			Where("? = ?", bun.Ident(pk), id)

		_, err = q.Exec(c)

		if err != nil {
			fmt.Println(q.String())
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err, "sql": q})
			return
		}

		c.JSON(http.StatusOK, newModel)
		return
	}
}
