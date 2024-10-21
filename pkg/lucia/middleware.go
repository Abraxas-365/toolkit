package lucia

import (
	"time"

	"github.com/Abraxas-365/toolkit/pkg/errors"
	"github.com/gofiber/fiber/v2"
)

const SessionCookieName = "auth_session"

// AuthMiddleware creates a middleware that handles session validation and authentication
type AuthMiddleware[U AuthUser] struct {
	service *AuthService[U]
}

// NewAuthMiddleware creates a new instance of AuthMiddleware
func NewAuthMiddleware[U AuthUser](service *AuthService[U]) *AuthMiddleware[U] {
	return &AuthMiddleware[U]{service: service}
}

// SessionMiddleware creates a middleware that validates the session
func (am *AuthMiddleware[U]) SessionMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the session ID from the cookie
		sessionID := c.Cookies(SessionCookieName)
		if sessionID == "" {
			// If no session ID is provided, continue without setting the session
			return c.Next()
		}

		// Validate the session
		session, err := am.service.GetSession(c.Context(), sessionID)
		if err != nil {
			// If there's an error, clear the invalid session cookie
			c.ClearCookie(SessionCookieName)

			// Check if it's a "not found" error and return ErrUnauthorized
			if errors.IsLuciaError(err) {
				luciaErr := err.(errors.LuciaError)
				if luciaErr.Type == "UserSessionNotFound" {
					return errors.ErrUnauthorized("Session not found")
				}
			}

			// For other errors, continue without setting the session
			return c.Next()
		}

		// If the session is valid, store it in the context for later use
		c.Locals("session", session)

		return c.Next()
	}
}

// RequireAuth is a middleware that ensures a valid session exists
func (am *AuthMiddleware[U]) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		session := GetSession(c)
		if session == nil {
			return errors.ErrUnauthorized("Authentication required")
		}
		return c.Next()
	}
}

// GetSession retrieves the validated session from the context
func GetSession(c *fiber.Ctx) *Session {
	session, ok := c.Locals("session").(*Session)
	if !ok {
		return nil
	}
	return session
}

// SetSessionCookie sets the session cookie
func SetSessionCookie(c *fiber.Ctx, session *Session) {
	c.Cookie(&fiber.Cookie{
		Name:     SessionCookieName,
		Value:    session.ID,
		Expires:  time.Unix(session.ExpiresAt, 0),
		HTTPOnly: true,
		Secure:   true, // Set to true if using HTTPS
		SameSite: "Lax",
	})
}

// ClearSessionCookie clears the session cookie
func ClearSessionCookie(c *fiber.Ctx) {
	c.ClearCookie(SessionCookieName)
}
