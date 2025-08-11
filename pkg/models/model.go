package models

import (
	"errors"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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
