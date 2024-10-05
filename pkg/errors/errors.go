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
