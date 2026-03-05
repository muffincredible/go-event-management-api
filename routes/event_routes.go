package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/muffincredible/go-event-management-api/handlers"
	"github.com/muffincredible/go-event-management-api/middleware"
)

func EventRoutes(app *fiber.App) {
	api := app.Group("/events", middleware.AuthRequired)
	
	api.Post("/", handlers.CreateEvent)           //etkinlik oluştur
	api.Post("/join/:id", handlers.JoinEvent)     //etkinliğe katıl
	api.Delete("/:id", handlers.DeleteEvent)      //etkinlik sil
	
	//herkese açık rota
	app.Get("/events", handlers.ListEvents)       //tüm etkinlikleri gör
	api.Put("/:id", handlers.UpdateEvent) 		  //etkinlik güncelle
	api.Delete("/leave/:id", handlers.LeaveEvent) //ayrılma
	api.Get("/my-events", handlers.GetMyEvents)   //katıldığım etkinlikler
}