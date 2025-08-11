package models

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/iteais/sdk/pkg"
)

type ModelAfterLoad interface {
	AfterLoad(c *gin.Context)
}

type ModelLastModified interface {
	LastModifiedField() string
}

func LoadModel[T interface{}](c *gin.Context, model T, errorMessages map[string]string) (T, map[string][]string) {
	if err := c.ShouldBindJSON(&model); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			out := make(map[string][]string, len(ve))
			for _, field := range ve {
				out[field.Field()] = append(out[field.Field()], getErrorMsg(field, errorMessages))
			}
			return model, out
		}
	}

	if afterLoadModel, ok := interface{}(model).(ModelAfterLoad); ok {
		afterLoadModel.AfterLoad(c)
		return afterLoadModel.(T), nil
	}

	return model, nil
}

// CallModelFunc
// Usage:
// model := models.Event{}
// CallModelFunc(&model, "SetCreator", "value")
func CallModelFunc(model interface{}, methodName string, args ...interface{}) {
	rm := reflect.ValueOf(model)
	method := rm.MethodByName(methodName)
	if method.IsValid() != false {

		in := make([]reflect.Value, len(args))
		for i, arg := range args {
			in[i] = reflect.ValueOf(arg)
		}
		method.Call(in)
	}
}

func GetImage(entity string, entityId string, traceId string, defaultImage string) string {

	params := &url.Values{}
	params.Add("limit", "1")
	params.Add("offset", "0")
	params.Add("sort[field]", "created_at")
	params.Add("sort[order]", "DESC")
	params.Add("filter[entity]", entity)
	params.Add("filter[entity_id]", entityId)

	var StorageResponse struct {
		Id      int    `json:"id"`
		Message string `json:"message"`
		Url     string `json:"url"`
	}

	resp := pkg.InternalFetch(pkg.InternalFetchConfig{
		Method:  "GET",
		Url:     os.Getenv("STORAGE_SERVER") + "/internal/storage/list?" + params.Encode(),
		TraceId: traceId,
	})

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusOK {
		err = json.Unmarshal(body, &StorageResponse)
		if err != nil {
			return defaultImage
		}
		return StorageResponse.Url
	}

	return defaultImage
}
