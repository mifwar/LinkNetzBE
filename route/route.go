package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/mifwar/LinkSavvyBE/handler"
	"github.com/mifwar/LinkSavvyBE/middleware"
)

func InitRoutes() *fiber.App {
	app := fiber.New()

	corsConfig := cors.Config{
		AllowOrigins:     "*",
		AllowCredentials: true,
	}

	app.Use(cors.New(corsConfig))
	app.Get("/api/token", handler.Token)
	app.Post("/auth/login", handler.Login)
	app.Post("/auth/register", handler.Register)
	app.Get("/auth/logout", handler.Logout)
	app.Get("/auth/google", handler.GoogleSignIn)
	app.Get("/auth/google/callback", handler.GoogleCallback)

	app.Use(middleware.VerifyUser)
	app.Static("/uploads", "./uploads")
	app.Get("/api/user", handler.User)
	app.Post("/add/category", handler.Category)
	app.Post("/add/tag", handler.Tag)
	app.Post("/add/url", handler.Url)

	app.Get("/categories", handler.Categories)
	app.Get("/tags", handler.Tags)

	return app
}
