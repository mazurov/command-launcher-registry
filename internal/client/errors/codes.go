package errors

import (
	"fmt"
	"net/http"
	"os"
)

// Exit codes for different error scenarios
const (
	ExitSuccess          = 0 // Success
	ExitGeneralError     = 1 // General error (network failure, server 500, unknown error)
	ExitInvalidArguments = 2 // Invalid arguments/usage (missing required flags, invalid format)
	ExitNotFound         = 3 // Resource not found (404)
	ExitConflict         = 4 // Conflict (409) - e.g., resource already exists
	ExitAuthError        = 5 // Authentication error (401)
	ExitPermissionDenied = 6 // Permission denied (403)
)

// ExitWithError prints error message and exits with appropriate code
func ExitWithError(err error, message string) {
	if message != "" {
		fmt.Fprintf(os.Stderr, "Error: %s: %v\n", message, err)
	} else {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
	os.Exit(ExitGeneralError)
}

// ExitWithCode prints error message and exits with specific code
func ExitWithCode(code int, message string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", message)
	os.Exit(code)
}

// MapHTTPStatusToExitCode maps HTTP status codes to exit codes
func MapHTTPStatusToExitCode(statusCode int) int {
	switch statusCode {
	case http.StatusUnauthorized:
		return ExitAuthError
	case http.StatusForbidden:
		return ExitPermissionDenied
	case http.StatusNotFound:
		return ExitNotFound
	case http.StatusConflict:
		return ExitConflict
	case http.StatusBadRequest:
		return ExitInvalidArguments
	default:
		if statusCode >= 400 && statusCode < 500 {
			return ExitInvalidArguments
		}
		return ExitGeneralError
	}
}

// HandleHTTPError handles HTTP error responses
func HandleHTTPError(statusCode int, message string) {
	code := MapHTTPStatusToExitCode(statusCode)

	// Add specific suggestions for auth errors
	if statusCode == http.StatusUnauthorized {
		message += ". Try running 'cola-regctl login' to authenticate"
	}

	ExitWithCode(code, message)
}
