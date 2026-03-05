package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/muffincredible/go-event-management-api/handlers"
)

func AuthRoutes(app *fiber.App) {
	app.Post("/register", handlers.Register)
	app.Post("/login", handlers.Login)
}