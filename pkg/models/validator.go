package models

import (
	"github.com/go-playground/validator/v10"
	"strings"
)

type HttpError struct {
	Error string `json:"error"`
}

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
