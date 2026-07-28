package service

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func ValidateStruct(s interface{}) error {
	if err := validate.Struct(s); err != nil {
		if validationErrs, ok := err.(validator.ValidationErrors); ok {
			var errMsgs []string
			for _, e := range validationErrs {
				rule := e.Tag()
				if e.Param() != "" {
					rule = fmt.Sprintf("%s:%s", e.Tag(), e.Param())
				}
				errMsgs = append(errMsgs, fmt.Sprintf("field '%s' failed on rule '%s'", e.Field(), rule))
			}
			return fmt.Errorf("%s", strings.Join(errMsgs, ", "))
		}
		return err
	}
	return nil
}
