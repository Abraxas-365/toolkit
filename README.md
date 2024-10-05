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
	"github.com/Abraxas-365/toolkit/pkg/errors"
	"github.com/Abraxas-365/toolkit/pkg/lucia"
	"github.com/Abraxas-365/toolkit/pkg/lucia/infrastructure"
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

//Exmpale of use

func main() {
	// Initialize database connection
	db, err := sqlx.Connect("postgres", "your_connection_string")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Initialize repository and service
	repo := infrastructure.NewPostgresRepository(db)
	luciaService := lucia.NewService(repo)

	app := fiber.New(fiber.Config{
		ErrorHandler: errors.ErrorHandler,
	})

	// Protected routes
	api := app.Group("/api")
	api.Use(lucia.SessionMiddleware(luciaService))
	api.Use(lucia.RequireAuth)

	api.Get("/profile", func(c *fiber.Ctx) error {
		session := lucia.GetSession(c)
		return c.JSON(fiber.Map{
			"message":    "Protected route",
			"user_id":    session.UserID,
			"session_id": session.ID,
		})
	})

	app.Listen(":3000")
}
```
