package main

import (
	"github.com/Abraxas-365/toolkit/pkg/errors"
	"github.com/Abraxas-365/toolkit/pkg/lucia"
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
	repo := lucia.NewPostgresRepository(db)
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
