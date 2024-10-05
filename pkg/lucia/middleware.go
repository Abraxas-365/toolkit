package lucia

import (
	"github.com/Abraxas-365/toolkit/pkg/errors"
	"github.com/gofiber/fiber/v2"
)

const SessionCookieName = "auth_session"

// SessionMiddleware creates a middleware that validates the session
func SessionMiddleware(service *Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get the session ID from the cookie
		sessionID := c.Cookies(SessionCookieName)
		if sessionID == "" {
			return errors.NewLuciaError("InvalidSessionId", "No session ID provided")
		}

		// Validate the session
		session, err := service.ValidateSession(c.Context(), sessionID)
		if err != nil {
			// If there's an error, it will be a LuciaError, so we can return it directly
			return err
		}

		// If the session is valid, store it in the context for later use
		c.Locals("session", session)

		return c.Next()
	}
}

// GetSession retrieves the validated session from the context
func GetSession(c *fiber.Ctx) *UserSession {
	session, ok := c.Locals("session").(*UserSession)
	if !ok {
		return nil
	}
	return session
}

// RequireAuth is a middleware that ensures a valid session exists
func RequireAuth(c *fiber.Ctx) error {
	session := GetSession(c)
	if session == nil {
		return errors.NewLuciaError("Unauthorized", "Authentication required")
	}
	return c.Next()
}
