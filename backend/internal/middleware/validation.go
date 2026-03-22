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

// ──────────────────────────────────────────────────────────────────────
// Phone Number Validation
// ──────────────────────────────────────────────────────────────────────

// Supported country codes for phone validation
var supportedCountryCodes = map[string]int{
	"MA": 212, // Morocco
	"FR": 33,  // France
	"ES": 34,  // Spain
	"DE": 49,  // Germany
	"IT": 39,  // Italy
	"GB": 44,  // United Kingdom
	"US": 1,   // United States/Canada
	"DZ": 213, // Algeria
	"TN": 216, // Tunisia
}

// PhoneValidationResult contains the validated phone number and metadata.
type PhoneValidationResult struct {
	IsValid       bool   `json:"is_valid"`
	OriginalPhone string `json:"original_phone"`
	Normalized    string `json:"normalized"`
	CountryCode   string `json:"country_code"`
	CountryName   string `json:"country_name"`
	Format        string `json:"format"` // "E164" or "local"
	Error         string `json:"error,omitempty"`
}

// ValidatePhoneNumber validates a phone number and returns normalized form.
// Supports:
// - Moroccan formats: +212 6XX XXX XXX, 06XX XXX XXX, 2126XXXXXXXX
// - E.164 format: +212612345678
// - International formats with country codes
func ValidatePhoneNumber(phone string) *PhoneValidationResult {
	result := &PhoneValidationResult{
		OriginalPhone: phone,
	}

	if phone == "" {
		result.Error = "phone number is required"
		return result
	}

	// Remove common formatting characters but keep +
	cleaned := strings.NewReplacer(
		" ", "", "-", "", "(", "", ")", "",
		".", "", ",", "", "", "",
	).Replace(phone)

	// Check if starts with +
	hasPlus := strings.HasPrefix(cleaned, "+")
	if hasPlus {
		cleaned = cleaned[1:]
	}

	// Must be numeric after cleanup
	if !isNumeric(cleaned) {
		result.Error = "phone number must contain only digits (and optional + prefix)"
		return result
	}

	// Detect country code and validate length
	// Try to match known country codes
	var countryCode int
	var countryName string

	// Check for 3-digit country codes first (212, 213, 216, etc.)
	if len(cleaned) >= 3 {
		prefix3 := cleaned[:3]
		if code, ok := map[string]int{"212": 212, "213": 213, "216": 216}[prefix3]; ok {
			countryCode = code
			cleaned = cleaned[3:]
			countryName = getCountryNameByCode(code)
		}
	}

	// Check for 2-digit country codes (33, 34, 49, etc.)
	if countryCode == 0 && len(cleaned) >= 2 {
		prefix2 := cleaned[:2]
		if code, ok := map[string]int{"33": 33, "34": 34, "49": 49, "39": 39, "44": 44, "1": 1}[prefix2]; ok {
			countryCode = code
			cleaned = cleaned[2:]
			countryName = getCountryNameByCode(code)
		}
	}

	// If no country code detected, assume Moroccan (212)
	if countryCode == 0 {
		// Handle local Moroccan format (06XXXXXXXX)
		if strings.HasPrefix(phone, "06") || strings.HasPrefix(phone, "07") || strings.HasPrefix(phone, "05") {
			countryCode = 212 // Morocco
			countryName = "Morocco"
			// Remove leading 0
			if strings.HasPrefix(cleaned, "0") {
				cleaned = cleaned[1:]
			}
		} else if hasPlus {
			// E.164 format with +, but unrecognized country code
			result.Error = "unsupported country code. Supported: +212 (Morocco), +33 (France), +34 (Spain), +49 (Germany), +39 (Italy), +44 (UK), +1 (US/CA), +213 (Algeria), +216 (Tunisia)"
			return result
		} else {
			// Default to Morocco if it looks like a valid mobile number
			if len(cleaned) >= 9 {
				countryCode = 212
				countryName = "Morocco"
			} else {
				result.Error = "invalid phone number length. Expected 9-10 digits for Moroccan numbers"
				return result
			}
		}
	}

	// Validate remaining number length
	// Moroccan numbers: 9 digits (6XX XXX XXX)
	if countryCode == 212 {
		if len(cleaned) != 9 {
			result.Error = "Moroccan phone number must be 9 digits (e.g., 06XX XXX XXX or +212 6XX XXX XXX)"
			return result
		}
		// Validate it starts with 6 or 7 (mobile)
		if !strings.HasPrefix(cleaned, "6") && !strings.HasPrefix(cleaned, "7") {
			result.Error = "Moroccan phone number must start with 6 or 7"
			return result
		}
	}

	// Build normalized E.164 format
	normalized := fmt.Sprintf("+%d%s", countryCode, cleaned)

	result.IsValid = true
	result.Normalized = normalized
	result.CountryCode = fmt.Sprintf("+%d", countryCode)
	result.CountryName = countryName
	result.Format = "E164"

	return result
}

// ValidateMoroccanPhoneNumber specifically validates Moroccan phone numbers.
// Supports: +212 6XX XXX XXX, 06XX XXX XXX, 2126XXXXXXXX, 6XXXXXXXX
func ValidateMoroccanPhoneNumber(phone string) *PhoneValidationResult {
	// First try general validation
	result := ValidatePhoneNumber(phone)

	// Then verify it's a Moroccan number
	if result.IsValid && result.CountryCode != "+212" {
		result.IsValid = false
		result.Error = "this field must be a valid Moroccan phone number (e.g., 06XX XXX XXX or +212 6XX XXX XXX)"
		result.Normalized = ""
	}

	return result
}

// isNumeric checks if a string contains only digits.
func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// getCountryNameByCode returns the country name for a given country code.
func getCountryNameByCode(code int) string {
	names := map[int]string{
		212: "Morocco",
		213: "Algeria",
		216: "Tunisia",
		33:  "France",
		34:  "Spain",
		49:  "Germany",
		39:  "Italy",
		44:  "United Kingdom",
		1:   "United States/Canada",
	}
	if name, ok := names[code]; ok {
		return name
	}
	return "Unknown"
}
