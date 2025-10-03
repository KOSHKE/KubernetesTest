package validation

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validate is an alias for validator.Validate to avoid naming conflicts
type Validate = validator.Validate

// New creates a new validator with common validation rules
func New() *Validate {
	v := validator.New()

	// Add custom validation rules here if needed
	// Example: v.RegisterValidation("custom_rule", customValidationFunc)

	return v
}

// FieldViolation represents a single field validation error
type FieldViolation struct {
	Field       string
	Description string
}

// FieldError represents a single field error that implements error
type FieldError struct {
	Field       string
	Description string
}

func (e FieldError) Error() string {
	return e.Field + " " + e.Description
}

// NewFieldError creates a new field error
func NewFieldError(field, description string) error {
	return FieldError{Field: field, Description: description}
}

// ExtractViolations extracts field validation violations from common validation errors
// Supports go-playground/validator errors and selected value object errors
func ExtractViolations(err error) ([]FieldViolation, bool) {
	var verrs validator.ValidationErrors
	if errors.As(err, &verrs) {
		violations := make([]FieldViolation, 0, len(verrs))
		for _, fe := range verrs {
			field := fe.Field()
			switch fe.Tag() {
			case "required":
				violations = append(violations, FieldViolation{Field: field, Description: "is required"})
			case "email":
				violations = append(violations, FieldViolation{Field: field, Description: "must be a valid email"})
			case "min":
				violations = append(violations, FieldViolation{Field: field, Description: fmt.Sprintf("must be at least %s characters", fe.Param())})
			case "max":
				violations = append(violations, FieldViolation{Field: field, Description: fmt.Sprintf("must be at most %s characters", fe.Param())})
			default:
				violations = append(violations, FieldViolation{Field: field, Description: "is invalid"})
			}
		}
		return violations, true
	}

	// Support direct FieldError from value objects or other layers
	var fe FieldError
	if errors.As(err, &fe) {
		return []FieldViolation{{Field: fe.Field, Description: fe.Description}}, true
	}

	return nil, false
}

// BuildMessage builds a concise, user-friendly message from violations list
func BuildMessage(violations []FieldViolation) string {
	if len(violations) == 0 {
		return "invalid request"
	}
	parts := make([]string, 0, len(violations))
	for _, v := range violations {
		parts = append(parts, fmt.Sprintf("%s %s", v.Field, v.Description))
	}
	return strings.Join(parts, ", ")
}
