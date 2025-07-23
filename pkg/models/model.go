package models

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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

	if afterLoadModel, ok := interface{}(model).(ModelAfterLoad); ok {
		afterLoadModel.AfterLoad()
		return afterLoadModel.(T), nil
	}

	return model, nil
}
