package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/mifwar/LinkSavvyBE/handler"
	_ "github.com/mifwar/LinkSavvyBE/middleware"
)

func InitRoutes() *fiber.App {
	app := fiber.New()

	corsConfig := cors.Config{
		AllowOrigins:     "*",
		AllowCredentials: true,
	}

	app.Use(cors.New(corsConfig))

	app.Post("/auth/login", handler.Login)
	app.Post("/auth/register", handler.Register)
	app.Get("/auth/logout", handler.Logout)
	app.Get("/auth/google", handler.GoogleSignIn)
	app.Get("/auth/google/callback", handler.GoogleCallback)
	app.Get("/api/token", handler.Token)
	app.Get("/api/user", handler.User)

	return app
}
