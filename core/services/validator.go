package services

import (
	"regexp"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func RegisterValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		idRegex := regexp.MustCompile(`^[A-Za-z0-9_]+$`)
		v.RegisterValidation("identifier", func(fl validator.FieldLevel) bool {
			return idRegex.MatchString(fl.Field().String())
		})
	}
}
