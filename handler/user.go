package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mifwar/LinkSavvyBE/auth"
	"github.com/mifwar/LinkSavvyBE/db"
	"github.com/mifwar/LinkSavvyBE/db/models"
)

func User(c *fiber.Ctx) error {

	userID := c.Locals("userID")

	userEmail := db.GetByUserID("email", userID.(float64))
	name := db.GetByUserID("full_name", userID.(float64))

	tokenString := auth.GenerateJWT(db.GetUserID(userEmail))

	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    tokenString,
		Expires:  time.Now().Add(time.Hour * 24),
		HTTPOnly: true,
		SameSite: "none",
	}

	c.Cookie(&cookie)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"email": userEmail,
		"name":  name,
	})
}

func Register(c *fiber.Ctx) error {
	newUser := models.NewUser{}

	if err := c.BodyParser(&newUser); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "failed to parse",
		})
	}

	newUser.Encrypt()

	if err := db.CreateUser(newUser, "email"); err != nil {
		if strings.Contains(err.Error(), `duplicate key value violates unique constraint "users_email_key"`) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "This email is already in use",
			})
		}
	}

	return c.JSON(newUser)
}
