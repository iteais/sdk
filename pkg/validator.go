package pkg

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"strings"
)

func getErrorMsg(fe validator.FieldError, messages map[string]string) string {

	if _, ok := messages[fe.Tag()]; ok {
		return strings.ReplaceAll("{attribute}", messages[fe.Tag()], fe.Tag())
	}

	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "lte":
		return "Should be less than " + fe.Param()
	case "gte":
		return "Should be greater than " + fe.Param()
	}
	return fe.Error()
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

	return model, nil
}
