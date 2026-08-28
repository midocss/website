// Package httpx contains small helpers shared by the HTTP handlers.
package httpx

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/midocss/website/pkg/apperr"
)

// BindJSON decodes and validates the request body, translating validator
// failures into the unified validation error format.
func BindJSON(c *gin.Context, target any) error {
	if err := c.ShouldBindJSON(target); err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			appErr := apperr.Validation("the submitted data is invalid")
			for _, fieldErr := range validationErrs {
				appErr = appErr.WithFields(apperr.FieldError{
					Field:   toSnakeCase(fieldErr.Field()),
					Message: describe(fieldErr),
				})
			}
			return appErr
		}
		return apperr.BadRequest("malformed JSON body").WithCause(err)
	}
	return nil
}

// ClientIP returns the caller IP without the port.
func ClientIP(c *gin.Context) string {
	return c.ClientIP()
}

func describe(fieldErr validator.FieldError) string {
	switch fieldErr.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s characters long", fieldErr.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters long", fieldErr.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", strings.ReplaceAll(fieldErr.Param(), " ", ", "))
	case "uuid", "uuid4":
		return "must be a valid UUID"
	default:
		return fmt.Sprintf("failed the %s validation rule", fieldErr.Tag())
	}
}

func toSnakeCase(field string) string {
	var b strings.Builder
	for i, r := range field {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(r + ('a' - 'A'))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
