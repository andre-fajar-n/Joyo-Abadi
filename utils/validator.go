package utils

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator instance
var Validate *validator.Validate

// InitValidator initializes the validator instance
func InitValidator() {
	Validate = validator.New()

	// Register custom validation tags
	registerCustomValidations()
}

// registerCustomValidations registers custom validation rules
func registerCustomValidations() {
	// Custom validation for safe text (alphanumeric with spaces, underscores, hyphens)
	// Useful for names, titles, descriptions that need to be safe but readable
	Validate.RegisterValidation("safe_text", func(fl validator.FieldLevel) bool {
		text := fl.Field().String()
		if len(text) == 0 {
			return false
		}
		// Allow letters, numbers, spaces, underscores, and hyphens
		for _, char := range text {
			if !((char >= 'a' && char <= 'z') ||
				(char >= 'A' && char <= 'Z') ||
				(char >= '0' && char <= '9') ||
				char == ' ' || char == '_' || char == '-') {
				return false
			}
		}
		return true
	})
}

// ValidateStruct validates a struct and returns user-friendly error messages
func ValidateStruct(s interface{}) []string {
	var errors []string

	err := Validate.Struct(s)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, formatValidationError(err))
		}
	}

	return errors
}

// formatValidationError converts validator errors to user-friendly messages
func formatValidationError(err validator.FieldError) string {
	field := strings.ToLower(strings.ReplaceAll(err.Field(), "_", " "))

	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", field, err.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", field, err.Param())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "safe_text":
		return fmt.Sprintf("%s can only contain letters, numbers, spaces, underscores, and hyphens", field)
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, err.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, err.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, err.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, err.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", field, err.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, err.Param())
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// ValidationError represents a validation error response
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// ValidateStructDetailed validates a struct and returns detailed error information
func ValidateStructDetailed(s interface{}) []ValidationError {
	var errors []ValidationError

	err := Validate.Struct(s)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, ValidationError{
				Field:   err.Field(),
				Message: formatValidationError(err),
				Value:   fmt.Sprintf("%v", err.Value()),
			})
		}
	}

	return errors
}
