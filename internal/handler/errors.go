package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func init() {
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

		if name == "-" {
			return ""
		}

		return name
	})
}

type APIError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func respondWithError(w http.ResponseWriter, r *http.Request, code int, message string, originalError error, details ...any) {
	// Log the original error with request context for debugging
	fmt.Printf("Api Error: Status=%d, Message=%s, Request=%s, Error=%v, Details=%v\n", code, message, r.URL.String(), originalError, details)

	responseBody := APIError{
		Status:  code,
		Message: message,
	}

	if len(details) > 0 {
		if len(details) == 1 {
			responseBody.Details = details[0]
		} else {
			responseBody.Details = details
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// Ensure Vary header is set when content negotiation might happen (even if just application/json now)
	w.Header().Set("Vary", "Accept")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(responseBody); err != nil {
		fmt.Printf("Error encoding error response: %v\n", err)
		http.Error(w, `{"status":500,"message":"Internal Server Error encoding error response"}`, http.StatusInternalServerError)
	}
}

// validationErrorResponse structures the validation error details.
type validationErrorResponse struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// formatValidationErrors converts validator errors into a user-friendly slice.
func formatValidationErrors(err error) []validationErrorResponse {
	var validationErrors []validationErrorResponse

	// Check if the error is actually validator.ValidationErrors
	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, err := range errs {
			// Use JSON field name from the tag name func we registered
			field := err.Field()
			message := fmt.Sprintf("failed on '%s' validation", err.Tag())

			switch err.Tag() {
			case "required":
				message = "is required"
			case "min":
				message = fmt.Sprintf("must be at least %s characters long", err.Param())
			case "max":
				message = fmt.Sprintf("must be at most %s characters long", err.Param())
			case "url":
				message = "must be a valid URL"
			}
			validationErrors = append(validationErrors, validationErrorResponse{
				Field:   field,
				Message: message,
			})
		}
	} else {
		// Handle non-validation errors if they somehow reach here
		// Or just return a generic error detail
		fmt.Printf("Warning: formatValidationErrors received non-validation error: %v\n", err)
	}

	return validationErrors
}
