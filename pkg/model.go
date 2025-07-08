package pkg

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"reflect"
)

type ModelAfterLoad interface {
	AfterLoad()
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

	value := reflect.ValueOf(model)
	modelType := reflect.Indirect(value).Type()
	modelValue := reflect.New(modelType)
	if afterLoadModel, ok := modelValue.Interface().(ModelAfterLoad); ok {
		afterLoadModel.AfterLoad()
	}

	return model, nil
}
