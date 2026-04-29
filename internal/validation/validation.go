package validation

import (
	"sync"

	"github.com/go-playground/validator/v10"

	"awesomeProject/service_errors"
)

var (
	validate     *validator.Validate
	validateOnce sync.Once
)

func getValidator() *validator.Validate {
	validateOnce.Do(func() {
		validate = validator.New()
	})
	return validate
}

// ValidateStruct validates any struct tagged with `validate:""` constraints.
// Returns a BadRequestError whose message lists all violated fields.
func ValidateStruct(s interface{}) error {
	if err := getValidator().Struct(s); err != nil {
		return service_errors.BadRequestError(err.Error())
	}
	return nil
}
