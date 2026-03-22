// Package middleware provides HTTP middleware for the MedConnect Oriental API.
package middleware

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ValidationError represents a single field validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// ErrorResponse is the standard error response structure.
type ErrorResponse struct {
	Error   string            `json:"error"`
	Details []ValidationError `json:"details,omitempty"`
}

// ValidationMiddleware provides structured validation error responses.
func ValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors in the context
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				if err.Type == gin.ErrorTypeBind {
					// Parse validation errors
					details := ParseValidationErrors(err.Err)

					c.JSON(400, ErrorResponse{
						Error:   "Validation failed",
						Details: details,
					})
					c.Abort()
					return
				}
			}
		}
	}
}

// ParseValidationErrors extracts structured validation errors.
// This is the exported version for use in handlers.
func ParseValidationErrors(err error) []ValidationError {
	var details []ValidationError

	// Check if it's a validator.ValidationErrors
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldErr := range validationErrors {
			details = append(details, ValidationError{
				Field:   strings.ToLower(fieldErr.Field()),
				Message: getErrorMessage(fieldErr),
				Code:    fieldErr.Tag(),
			})
		}
		return details
	}

	// Check for JSON unmarshal type errors (e.g., invalid UUID format)
	if strings.Contains(err.Error(), "json:") || strings.Contains(err.Error(), "parsing") {
		details = append(details, ValidationError{
			Field:   "request_body",
			Message: "Invalid JSON format or data type",
			Code:    "invalid_json",
		})
		return details
	}

	// Generic error
	details = append(details, ValidationError{
		Field:   "request",
		Message: err.Error(),
		Code:    "invalid_request",
	})

	return details
}

// getErrorMessage returns a user-friendly error message for a validation error.
func getErrorMessage(fe validator.FieldError) string {
	field := strings.ToLower(fe.Field())

	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		if fe.Type().Kind() == reflect.String {
			return fmt.Sprintf("%s must be at least %s characters", field, fe.Param())
		}
		return fmt.Sprintf("%s must be at least %s", field, fe.Param())
	case "max":
		if fe.Type().Kind() == reflect.String {
			return fmt.Sprintf("%s must be at most %s characters", field, fe.Param())
		}
		return fmt.Sprintf("%s must be at most %s", field, fe.Param())
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, fe.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, fe.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, fe.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, fe.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, fe.Param())
	case "alphanum":
		return fmt.Sprintf("%s must contain only alphanumeric characters", field)
	case "numeric":
		return fmt.Sprintf("%s must be a number", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "uri":
		return fmt.Sprintf("%s must be a valid URI", field)
	case "datetime":
		return fmt.Sprintf("%s must be a valid datetime in format %s", field, fe.Param())
	default:
		return fmt.Sprintf("%s failed validation: %s", field, fe.Tag())
	}
}

// ValidateStruct validates a struct and returns structured errors.
func ValidateStruct(v *validator.Validate, s interface{}) []ValidationError {
	var details []ValidationError

	if err := v.Struct(s); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldErr := range validationErrors {
				details = append(details, ValidationError{
					Field:   strings.ToLower(fieldErr.Field()),
					Message: getErrorMessage(fieldErr),
					Code:    fieldErr.Tag(),
				})
			}
		}
	}

	return details
}

// SanitizeInput removes potentially dangerous characters from input strings.
func SanitizeInput(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Remove control characters except newlines and tabs
	re := regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)
	input = re.ReplaceAllString(input, "")

	// Trim whitespace
	input = strings.TrimSpace(input)

	return input
}

// SanitizeStruct sanitizes all string fields in a struct.
func SanitizeStruct(s interface{}) {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanSet() {
			continue
		}

		if field.Kind() == reflect.String {
			sanitized := SanitizeInput(field.String())
			field.SetString(sanitized)
		} else if field.Kind() == reflect.Ptr && field.Type().Elem().Kind() == reflect.String {
			if !field.IsNil() {
				sanitized := SanitizeInput(field.Elem().String())
				field.Elem().SetString(sanitized)
			}
		}
	}
}

// BindAndValidate binds JSON to struct, sanitizes input, and validates.
// Returns structured validation errors if validation fails.
func BindAndValidate(c *gin.Context, v *validator.Validate, dest interface{}) *ErrorResponse {
	// Bind JSON
	if err := c.ShouldBindJSON(dest); err != nil {
		return &ErrorResponse{
			Error:   "Invalid request body",
			Details: ParseValidationErrors(err),
		}
	}

	// Sanitize input
	SanitizeStruct(dest)

	// Validate
	if details := ValidateStruct(v, dest); len(details) > 0 {
		return &ErrorResponse{
			Error:   "Validation failed",
			Details: details,
		}
	}

	return nil
}

// RespondWithError sends a structured error response.
func RespondWithError(c *gin.Context, statusCode int, message string, details ...ValidationError) {
	response := ErrorResponse{
		Error: message,
	}

	if len(details) > 0 {
		response.Details = details
	}

	c.JSON(statusCode, response)
}

// RespondWithValidationError sends a 400 response with validation details.
func RespondWithValidationError(c *gin.Context, details []ValidationError) {
	c.JSON(400, ErrorResponse{
		Error:   "Validation failed",
		Details: details,
	})
}

// ParseJSONErrorDetails extracts error details from a JSON error response.
func ParseJSONErrorDetails(data []byte) (*ErrorResponse, error) {
	var response ErrorResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}
	return &response, nil
}
