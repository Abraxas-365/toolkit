package errors

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
)

// ApiError represents custom API errors
type ApiError struct {
	Type    string
	Message string
}

// Error implements the error interface
func (e ApiError) Error() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// NewApiError creates a new ApiError
func NewApiError(errType string, message string) ApiError {
	return ApiError{Type: errType, Message: message}
}

// Error types
var (
	ErrParse              = func(msg string) ApiError { return NewApiError("ParseError", msg) }
	ErrUnexpected         = func(msg string) ApiError { return NewApiError("UnexpectedError", msg) }
	ErrDatabase           = func(msg string) ApiError { return NewApiError("DatabaseError", msg) }
	ErrNotFound           = func(msg string) ApiError { return NewApiError("NotFound", msg) }
	ErrBadRequest         = func(msg string) ApiError { return NewApiError("BadRequest", msg) }
	ErrForbidden          = func(msg string) ApiError { return NewApiError("Forbidden", msg) }
	ErrUnauthorized       = func(msg string) ApiError { return NewApiError("Unauthorized", msg) }
	ErrConflict           = func(msg string) ApiError { return NewApiError("Conflict", msg) }
	ErrServiceUnavailable = func(msg string) ApiError { return NewApiError("ServiceUnavailable", msg) }
)

// LuciaError represents errors from the Lucia authentication library
type LuciaError struct {
	Type    string
	Message string
}

// Error implements the error interface
func (e LuciaError) Error() string {
	return fmt.Sprintf("Lucia error - %s: %s", e.Type, e.Message)
}

// NewLuciaError creates a new LuciaError
func NewLuciaError(errType string, message string) LuciaError {
	return LuciaError{Type: errType, Message: message}
}

// ErrorHandler is a custom error handler for Fiber
func ErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	if e, ok := errors.Cause(err).(ApiError); ok {
		code, message = handleApiError(e)
	} else if le, ok := errors.Cause(err).(LuciaError); ok {
		code, message = handleLuciaError(le)
	}

	return c.Status(code).JSON(fiber.Map{
		"error": message,
	})
}

// handleApiError determines the appropriate HTTP status code and message for ApiErrors
func handleApiError(e ApiError) (int, string) {
	switch e.Type {
	case "ParseError", "UnexpectedError", "DatabaseError":
		return fiber.StatusInternalServerError, e.Message
	case "NotFound":
		return fiber.StatusNotFound, e.Message
	case "BadRequest":
		return fiber.StatusBadRequest, e.Message
	case "Forbidden":
		return fiber.StatusForbidden, e.Message
	case "Unauthorized":
		return fiber.StatusUnauthorized, e.Message
	case "Conflict":
		return fiber.StatusConflict, e.Message
	case "ServiceUnavailable":
		return fiber.StatusServiceUnavailable, e.Message
	default:
		return fiber.StatusInternalServerError, e.Message
	}
}

// handleLuciaError determines the appropriate HTTP status code and message for LuciaErrors
func handleLuciaError(le LuciaError) (int, string) {
	switch le.Type {
	case "DatabaseConnectionError", "DatabaseQueryError", "UserSessionTableNotExist", "AuthUserTableNotExist", "SessionCreationFailed", "SessionDeletionFailed", "UserCreationFailed", "UserUpdateFailed", "EncryptionError", "DecryptionError", "ConfigurationError", "UnexpectedError":
		return fiber.StatusInternalServerError, le.Message
	case "UserSessionNotFound":
		return fiber.StatusNotFound, le.Message
	case "InvalidSessionId":
		return fiber.StatusBadRequest, le.Message
	case "SessionExpired", "InvalidCredentials", "InvalidToken", "TokenExpired":
		return fiber.StatusUnauthorized, le.Message
	case "DuplicateUserError":
		return fiber.StatusConflict, le.Message
	default:
		return fiber.StatusInternalServerError, le.Message
	}
}

func IsParseError(err error) bool {
	e, ok := errors.Cause(err).(ApiError)
	return ok && e.Type == "ParseError"
}

// IsUnexpectedError checks if the error is an UnexpectedError
func IsUnexpectedError(err error) bool {
	e, ok := errors.Cause(err).(ApiError)
	return ok && e.Type == "UnexpectedError"
}

// IsDatabaseError checks if the error is a DatabaseError
func IsDatabaseError(err error) bool {
	e, ok := errors.Cause(err).(ApiError)
	return ok && e.Type == "DatabaseError"
}

// IsNotFound checks if the error is a NotFound error
func IsNotFound(err error) bool {
	e, ok := errors.Cause(err).(ApiError)
	return ok && e.Type == "NotFound"
}

// IsBadRequest checks if the error is a BadRequest error
func IsBadRequest(err error) bool {
	e, ok := errors.Cause(err).(ApiError)
	return ok && e.Type == "BadRequest"
}

// IsForbidden checks if the error is a Forbidden error
func IsForbidden(err error) bool {
	e, ok := errors.Cause(err).(ApiError)
	return ok && e.Type == "Forbidden"
}

// IsUnauthorized checks if the error is an Unauthorized error
func IsUnauthorized(err error) bool {
	e, ok := errors.Cause(err).(ApiError)
	return ok && e.Type == "Unauthorized"
}

// IsConflict checks if the error is a Conflict error
func IsConflict(err error) bool {
	e, ok := errors.Cause(err).(ApiError)
	return ok && e.Type == "Conflict"
}

// IsServiceUnavailable checks if the error is a ServiceUnavailable error
func IsServiceUnavailable(err error) bool {
	e, ok := errors.Cause(err).(ApiError)
	return ok && e.Type == "ServiceUnavailable"
}

// IsLuciaError checks if the error is a LuciaError
func IsLuciaError(err error) bool {
	_, ok := errors.Cause(err).(LuciaError)
	return ok
}

// Additional helper functions for specific Lucia error types

// IsLuciaDatabaseError checks if the error is a Lucia database-related error
func IsLuciaDatabaseError(err error) bool {
	le, ok := errors.Cause(err).(LuciaError)
	return ok && (le.Type == "DatabaseConnectionError" || le.Type == "DatabaseQueryError")
}

// IsLuciaSessionError checks if the error is a Lucia session-related error
func IsLuciaSessionError(err error) bool {
	le, ok := errors.Cause(err).(LuciaError)
	return ok && (le.Type == "UserSessionNotFound" || le.Type == "InvalidSessionId" || le.Type == "SessionExpired")
}

// IsLuciaAuthError checks if the error is a Lucia authentication-related error
func IsLuciaAuthError(err error) bool {
	le, ok := errors.Cause(err).(LuciaError)
	return ok && (le.Type == "InvalidCredentials" || le.Type == "InvalidToken" || le.Type == "TokenExpired")
}

// IsLuciaDuplicateUserError checks if the error is a Lucia duplicate user error
func IsLuciaDuplicateUserError(err error) bool {
	le, ok := errors.Cause(err).(LuciaError)
	return ok && le.Type == "DuplicateUserError"
}
