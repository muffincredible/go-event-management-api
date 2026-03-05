package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/muffincredible/go-event-management-api/configs"
	"github.com/muffincredible/go-event-management-api/routes"
	"log"
)

func main() {
	app := fiber.New()

	configs.ConnectDB()

	routes.AuthRoutes(app)
	routes.EventRoutes(app)

	//test
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Sistem ve Veritabanı hazır.")
	})

	log.Fatal(app.Listen(":3000"))
}