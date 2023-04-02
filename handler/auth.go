package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/mifwar/LinkSavvyBE/auth"
	"github.com/mifwar/LinkSavvyBE/db"
	"github.com/mifwar/LinkSavvyBE/db/models"
)

func Token(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	authKey := strings.Replace(authHeader, "Bearer ", "", -1)

	localAuth := os.Getenv("AUTH_KEY")

	if authKey != localAuth {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "not authorized",
		})
	}

	cookie := c.Cookies("jwt")

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "ok",
		"token":   cookie,
	})
}

func Login(c *fiber.Ctx) error {

	user := models.User{}

	if err := c.BodyParser(&user); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "failed to parse",
		})
	}

	if err := auth.ValidateEmail(user.Email); err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"message": "Email is not registered",
		})
	}

	if err := auth.ValidateLoginMethod(user.Email, "email"); err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	if err := auth.ValidatePassword(user.Email, user.Password); err != nil {
		if strings.Contains(err.Error(), "hashedPassword is not the hash of the given password") {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"message": "Invalid password",
			})
		}

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Unknown error",
		})
	}

	tokenString := auth.GenerateJWT(db.GetUserID(user.Email))

	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    tokenString,
		Expires:  time.Now().Add(time.Hour * 24),
		HTTPOnly: true,
		SameSite: "none",
	}

	c.Cookie(&cookie)

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"message": "authorized",
		"token":   tokenString,
	})
}

func Logout(c *fiber.Ctx) error {
	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		SameSite: "none",
	}

	c.Cookie(&cookie)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "successfully logged out",
	})
}

func GoogleSignIn(c *fiber.Ctx) error {

	state := uuid.New().String()

	session, err := Store.Get(c)

	if session == nil {
		log.Fatal("Session not found")
	}

	if err != nil {
		log.Fatal(err)
	}

	session.Set("state", state)

	if err := session.Save(); err != nil {
		log.Fatal(err)
	}

	url := auth.GoogleOAuthConfig.AuthCodeURL(state)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"url": url,
	})
}

func GoogleCallback(c *fiber.Ctx) error {
	session, err := Store.Get(c)
	if err != nil {
		log.Fatal(err)
	}

	queryState := c.Query("state")
	sessionState := session.Get("state")

	if sessionState != queryState {
		c.Redirect(os.Getenv("FRONTEND_URL"))
		return nil

	}

	code := c.Query("code")
	token, err := auth.GoogleOAuthConfig.Exchange(c.Context(), code)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "token exchange failed",
		})
	}

	client := auth.GoogleOAuthConfig.Client(c.Context(), token)
	response, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "failed to get user info",
		})
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "failed to proceed response body",
		})
	}

	var userInfo map[string]interface{}
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "failed to extract response body",
		})
	}

	name := userInfo["name"].(string)
	email := userInfo["email"].(string)

	frontEndURL := os.Getenv("FRONTEND_URL")

	//register user if email is not registered yet
	if err := auth.ValidateEmail(email); err != nil {
		password := uuid.New().String()
		newUser := models.NewUser{FullName: name, Email: email, Password: password}
		newUser.Encrypt()

		db.CreateUser(newUser, "google")
	} else {
		if errLoginMethod := auth.ValidateLoginMethod(email, "google"); errLoginMethod != nil {
			url := fmt.Sprintf("%s/auth/wrongMethod", frontEndURL)
			c.Redirect(url)
			return nil
		}
	}

	tokenString := auth.GenerateJWT(db.GetUserID(email))

	cookie := fiber.Cookie{
		Name:     "jwt",
		Value:    tokenString,
		Expires:  time.Now().Add(time.Hour * 24),
		HTTPOnly: true,
		SameSite: "none",
	}

	c.Cookie(&cookie)

	c.Redirect(frontEndURL)
	return nil
}
