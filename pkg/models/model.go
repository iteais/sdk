package models

import (
	"errors"
	"reflect"
	"strings"

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

// GetModelFields returns a set of allowed field names for the model type T.
// It uses reflection to get the field names from the model struct.
// If the model is a pointer, it dereferences it to get the actual struct type.
// <b>Attention</b> it supports only the bun tag to get the column name for each field.
func GetModelFields[T any]() map[string]struct{} {
	var allowed = make(map[string]struct{})
	var t T
	typ := reflect.TypeOf(t)
	// If T is a pointer, get the element type
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() == reflect.Struct {
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			// Use the bun tag if present, otherwise the struct field name
			col := field.Tag.Get("bun")
			if col == "" || col == "-" {
				col = field.Name
			} else {
				// bun tag may have options, take only the column name
				col = strings.Split(col, ",")[0]
			}
			allowed[col] = struct{}{}
		}
	}
	return allowed
}

// GetAllProps returns a set of allowed field names for the model type T.
func GetAllProps(t any) []string {
	var allowed = make([]string, 0)
	//var t T
	typ := reflect.TypeOf(t)
	// If T is a pointer, get the element type
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() == reflect.Struct {
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)

			allowed = append(allowed, field.Name)

		}
	}
	return allowed
}
