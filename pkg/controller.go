package pkg

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iteais/sdk/pkg/models"
	"github.com/iteais/sdk/pkg/utils"
	"github.com/uptrace/bun"
)

func ListAction[T interface{}](postFindFuncs ...func(*gin.Context, *[]T)) func(c *gin.Context) {
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
			allowedFields := models.GetModelFields[T]()
			requestedFields := strings.Split(fields, ",")
			validFields := make([]string, 0, len(requestedFields))
			for _, f := range requestedFields {
				f = strings.TrimSpace(f)
				if _, ok := allowedFields[f]; ok {
					validFields = append(validFields, f)
				}
			}
			if len(validFields) > 0 {
				query = query.Column(validFields...)
			}
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
				relation = utils.ToUpperCamelCase(strings.TrimSpace(relation))
				query = query.Relation(relation)
			}
		}

		ApplyFilter[T](c, query)

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			appendLastModifiedHeader[T](c)
		}()

		fmt.Println(query.String())

		count, err := query.ScanAndCount(context.Background())

		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Header("X-Total-Count", fmt.Sprintf("%d", count))
		c.Header("x-pagination-per-page", fmt.Sprintf("%d", perPage))

		xppc := 1
		calcXppc := math.Ceil(float64(count) / float64(perPage))
		if calcXppc > 0 {
			xppc = int(calcXppc)
		}

		c.Header("x-pagination-page-count", fmt.Sprintf("%d", xppc))

		//x-pagination-current-page
		c.Header("x-pagination-current-page", fmt.Sprintf("%d", page))

		for _, f := range postFindFuncs {
			f(c, &modelsArray)
		}

		wg.Wait()
		c.JSON(200, gin.H{"data": modelsArray})
	}
}

func appendLastModifiedHeader[T interface{}](c *gin.Context) {
	model := new(T)
	if found, ok := interface{}(model).(models.ModelLastModified); ok {

		field := found.LastModifiedField()

		var timeString string

		query := App.Db.NewSelect().
			Model(model).ColumnExpr(field).Order(field + " DESC").Limit(1)

		ApplyFilter[T](c, query)

		err := query.Scan(context.Background(), &timeString)
		if err == nil && timeString != "" {
			layout := "2006-01-02T15:04:05.999999Z"
			dt, err := time.Parse(layout, timeString)
			if err == nil {
				format := dt.Format(time.RFC1123)
				c.Header("Last-Modified", format)
			}
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

func CreateAction[T interface{}]() func(*gin.Context) {
	return func(c *gin.Context) {

		modelType := new(T)

		model, loadErrors := models.LoadModel[T](c, *modelType, make(map[string]string))

		if len(loadErrors) > 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"errors": loadErrors})
			return
		}

		query := App.Db.NewInsert().Model(&model)

		App.Log.Info(query.String())

		_, err := query.Exec(context.Background())

		errMsg := ""

		if err != nil {
			errMsg = err.Error()
		}

		c.JSON(http.StatusCreated, gin.H{"data": model, "error": errMsg})
		return
	}
}

func GetByField[T interface{}](filterFiled string) gin.HandlerFunc {
	return func(c *gin.Context) {

		id := c.Param("id")
		fields := c.Query("fields")

		api := new(T)
		query := App.Db.NewSelect().
			Model(api).
			Where("? = ?", bun.Ident(filterFiled), id)

		allowedColumns := models.GetAllProps[api]()

		var safeColumns []string
		if fields != "" {
			// Split by comma, trim whitespace, validate against allowedColumns
			requested := strings.Split(fields, ",")
			for _, col := range requested {
				col = strings.TrimSpace(col)
				for _, allowed := range allowedColumns {
					if col == allowed {
						safeColumns = append(safeColumns, col)
						break
					}
				}
			}
			if len(safeColumns) > 0 {
				query = query.Column(safeColumns...)
			}
		}

		App.GetRequestLogger(c).Info(query.String())

		count, err := query.ScanAndCount(c)

		if count < 1 {
			c.JSON(http.StatusNotFound, gin.H{})
			return
		}

		c.JSON(200, gin.H{"data": api, "error": err, "cnt": count})
	}
}
