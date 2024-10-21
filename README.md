# Go Toolkit

This toolkit is a collection of reusable packages and utilities designed to streamline the development of Go projects. It provides common functionalities and best practices that can be easily integrated into various Go applications.

## Features

- **Error Handling**: Custom error types and a centralized error handler for consistent error management across your projects.
- **Lucia Authentication**: A modular authentication system that can be easily integrated into web applications.
- **Database Utilities**: Helper functions and structures for database operations (currently supports PostgreSQL).

## Usage
```go
package main

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/Abraxas-365/toolkit/pkg/errors"
	"github.com/Abraxas-365/toolkit/pkg/lucia"
	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq"
)

type User struct {
	ID         string
	Email      string
	Name       string
	ProviderID string
	Provider   string
}

func (u *User) GetID() string {
	return u.ID
}

func main() {
	// Initialize in-memory user store
	authUserStore := NewInMemoryUserStore()
	sessionStore := NewInMemorySessionStore()
	// Initialize auth service
	authService := lucia.NewAuthService[*User](authUserStore, sessionStore)

	// Initialize Google OAuth provider
	googleProvider := lucia.NewGoogleProvider(
		os.Getenv("GOOGLE_CLIENT_ID"),
		os.Getenv("GOOGLE_CLIENT_SECRET"),
		os.Getenv("GOOGLE_REDIRECT_URI"),
	)
	authService.RegisterProvider("google", googleProvider)

	// Initialize auth middleware
	authMiddleware := lucia.NewAuthMiddleware(authService)

	app := fiber.New(fiber.Config{
		ErrorHandler: errors.ErrorHandler,
	})

	// Public routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Welcome to the public area!")
	})

	// Apply session middleware to all routes
	app.Use(authMiddleware.SessionMiddleware())

	// Protected routes
	api := app.Group("/api")
	api.Use(authMiddleware.RequireAuth())

	api.Get("/profile", func(c *fiber.Ctx) error {
		session := lucia.GetSession(c)
		return c.JSON(fiber.Map{
			"message":    "Protected route",
			"user_id":    session.UserID,
			"session_id": session.ID,
		})
	})

	// New route to get user info
	api.Get("/user", func(c *fiber.Ctx) error {
		session := lucia.GetSession(c)
		user, err := authUserStore.GetUserByID(c.Context(), session.UserID)
		if err != nil {
			return err
		}
		return c.JSON(fiber.Map{
			"id":       user.ID,
			"email":    user.Email,
			"name":     user.Name,
			"provider": user.Provider,
		})
	})

	// Google OAuth routes
	app.Get("/login/google", func(c *fiber.Ctx) error {
		authURL, state, err := authService.GetAuthURL("google")
		if err != nil {
			return err
		}
		c.Cookie(&fiber.Cookie{
			Name:     "oauth_state",
			Value:    state,
			HTTPOnly: true,
			Secure:   true,
		})
		return c.Redirect(authURL)
	})

	app.Get("/login/google/callback", func(c *fiber.Ctx) error {
		state := c.Cookies("oauth_state")
		if state == "" || state != c.Query("state") {
			return errors.ErrUnauthorized("Invalid state")
		}

		code := c.Query("code")
		if code == "" {
			return errors.ErrBadRequest("Missing code")
		}

		session, err := authService.HandleCallback(c.Context(), "google", code)
		if err != nil {
			return err
		}

		lucia.SetSessionCookie(c, session)
		return c.Redirect("/api/profile")
	})

	// Logout route
	app.Post("/logout", func(c *fiber.Ctx) error {
		session := lucia.GetSession(c)
		if session != nil {
			if err := authService.DeleteSession(c.Context(), session.ID); err != nil {
				return err
			}
		}
		lucia.ClearSessionCookie(c)
		return c.SendString("Logged out successfully")
	})

	app.Listen(":3000")
}

// InMemoryUserStore is a simple in-memory implementation of UserStore for testing
type InMemoryUserStore struct {
	users map[string]*User
	mu    sync.RWMutex
}

func NewInMemoryUserStore() *InMemoryUserStore {
	return &InMemoryUserStore{
		users: make(map[string]*User),
	}
}

func (s *InMemoryUserStore) GetUserByProviderID(ctx context.Context, provider, providerID string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.Provider == provider && user.ProviderID == providerID {
			return user, nil
		}
	}

	return nil, errors.ErrNotFound("User not found")
}

func (s *InMemoryUserStore) CreateUser(ctx context.Context, userInfo *lucia.UserInfo) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	user := &User{
		ID:         lucia.GenerateID(),
		Email:      userInfo.Email,
		Name:       userInfo.Name,
		ProviderID: userInfo.ID,
		Provider:   userInfo.Provider,
	}

	if _, exists := s.users[user.ID]; exists {
		return nil, errors.ErrConflict("User already exists")
	}

	s.users[user.ID] = user
	return user, nil
}

func (s *InMemoryUserStore) GetUserByID(ctx context.Context, userID string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[userID]
	if !exists {
		return nil, errors.ErrNotFound("User not found")
	}

	return user, nil
}

// InMemorySessionStore is a simple in-memory implementation of SessionStore for testing
type InMemorySessionStore struct {
	sessions map[string]*lucia.Session
	mu       sync.RWMutex
}

func NewInMemorySessionStore() *InMemorySessionStore {
	return &InMemorySessionStore{
		sessions: make(map[string]*lucia.Session),
	}
}

func (s *InMemorySessionStore) CreateSession(ctx context.Context, session *lucia.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sessions[session.ID]; exists {
		return errors.ErrConflict("Session already exists")
	}

	s.sessions[session.ID] = session
	return nil
}

func (s *InMemorySessionStore) GetSession(ctx context.Context, sessionID string) (*lucia.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return nil, errors.ErrNotFound("Session not found")
	}

	// Check if the session has expired
	if time.Now().Unix() > session.ExpiresAt {
		// Remove the expired session
		delete(s.sessions, sessionID)
		return nil, errors.ErrNotFound("Session expired")
	}

	return session, nil
}

func (s *InMemorySessionStore) DeleteSession(ctx context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sessions[sessionID]; !exists {
		return errors.ErrNotFound("Session not found")
	}

	delete(s.sessions, sessionID)
	return nil
}
```
