package validation

import (
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
