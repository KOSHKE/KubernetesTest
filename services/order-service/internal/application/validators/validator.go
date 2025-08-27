package validators

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/kubernetestest/ecommerce-platform/services/order-service/internal/application/dto"
)

// Validator provides centralized validation for DTOs
type Validator struct {
	validate *validator.Validate
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	v := validator.New()
	dto.RegisterCustomValidators(v)

	return &Validator{
		validate: v,
	}
}

// Validate validates a struct and returns human-readable error messages
func (v *Validator) Validate(req interface{}) error {
	if err := v.validate.Struct(req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			var errorMessages []string
			for _, fieldError := range validationErrors {
				errorMessages = append(errorMessages, v.formatFieldError(fieldError))
			}
			return fmt.Errorf("validation failed: %s", strings.Join(errorMessages, "; "))
		}
		return fmt.Errorf("validation failed: %w", err)
	}
	return nil
}

// formatFieldError formats a single field validation error
func (v *Validator) formatFieldError(fieldError validator.FieldError) string {
	field := fieldError.Field()
	tag := fieldError.Tag()
	param := fieldError.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("field '%s' is required", field)
	case "min":
		return fmt.Sprintf("field '%s' must have minimum length of %s", field, param)
	case "max":
		return fmt.Sprintf("field '%s' must have maximum length of %s", field, param)
	case "len":
		return fmt.Sprintf("field '%s' must have exact length of %s", field, param)
	case "gt":
		return fmt.Sprintf("field '%s' must be greater than %s", field, param)
	case "gte":
		return fmt.Sprintf("field '%s' must be greater than or equal to %s", field, param)
	case "lt":
		return fmt.Sprintf("field '%s' must be less than %s", field, param)
	case "lte":
		return fmt.Sprintf("field '%s' must be less than or equal to %s", field, param)
	case "oneof":
		return fmt.Sprintf("field '%s' must be one of: %s", field, param)
	case "valid_order_status":
		return fmt.Sprintf("field '%s' must be a valid order status", field)
	case "email":
		return fmt.Sprintf("field '%s' must be a valid email address", field)
	case "uuid":
		return fmt.Sprintf("field '%s' must be a valid UUID", field)
	case "dive":
		return fmt.Sprintf("field '%s' contains invalid items", field)
	default:
		return fmt.Sprintf("field '%s' failed validation: %s", field, tag)
	}
}

// ValidateSlice validates a slice of structs
func (v *Validator) ValidateSlice(slice interface{}) error {
	return v.validate.Var(slice, "dive")
}

// ValidateField validates a single field
func (v *Validator) ValidateField(field interface{}, tag string) error {
	return v.validate.Var(field, tag)
}
