package service

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"

	"guangjiapps/gin/internal/domain"
)

var validate = func() *validator.Validate {
	v := validator.New()
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	return v
}()

type ValidationError struct {
	Details map[string][]string
}

func (v *ValidationError) Error() string {
	var parts []string
	for field, msgs := range v.Details {
		for _, msg := range msgs {
			parts = append(parts, fmt.Sprintf("%s: %s", field, msg))
		}
	}
	return strings.Join(parts, ", ")
}

func NewValidationError(details map[string][]string) *ValidationError {
	return &ValidationError{Details: details}
}

func FormatValidatorErrors(err error) map[string][]string {
	details := make(map[string][]string)
	var validationErrs validator.ValidationErrors
	if ve, ok := err.(validator.ValidationErrors); ok {
		validationErrs = ve
	}
	if len(validationErrs) > 0 {
		for _, e := range validationErrs {
			field := e.Field()
			label := domain.GetFieldLabel(e.StructNamespace())
			if label == "" {
				label = domain.GetFieldLabel(e.StructField())
			}
			if label == "" {
				label = field
			}

			rule := e.Tag()
			if e.Param() != "" {
				rule = fmt.Sprintf("%s:%s", e.Tag(), e.Param())
			}

			var msg string
			switch e.Tag() {
			case "required":
				msg = fmt.Sprintf("%s wajib diisi", label)
			case "email":
				msg = fmt.Sprintf("%s harus berupa email yang valid (email)", label)
			case "min":
				msg = fmt.Sprintf("%s kurang dari nilai minimum %s (min)", label, e.Param())
			case "max":
				msg = fmt.Sprintf("%s melebihi nilai maksimum %s (max:%s)", label, e.Param(), e.Param())
			case "gte":
				msg = fmt.Sprintf("%s kurang dari nilai minimum %s", label, e.Param())
			case "lte":
				msg = fmt.Sprintf("%s melebihi nilai maksimum %s", label, e.Param())
			case "gtefield":
				targetLabel := domain.GetFieldLabel(e.Param())
				if targetLabel == "" {
					targetLabel = e.Param()
				}
				msg = fmt.Sprintf("%s tidak boleh lebih awal dari %s", label, targetLabel)
			case "ltefield":
				targetLabel := domain.GetFieldLabel(e.Param())
				if targetLabel == "" {
					targetLabel = e.Param()
				}
				msg = fmt.Sprintf("%s tidak boleh lebih lambat dari %s", label, targetLabel)
			default:
				msg = fmt.Sprintf("%s gagal pada aturan %s", label, rule)
			}
			details[field] = append(details[field], msg)
		}
	}
	return details
}

func ValidateStruct(s interface{}) error {
	if err := validate.Struct(s); err != nil {
		details := FormatValidatorErrors(err)
		if len(details) > 0 {
			return &ValidationError{Details: details}
		}
		return err
	}
	return nil
}
