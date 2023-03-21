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
	"github.com/gofiber/fiber/v2/middleware/session"

	"github.com/google/uuid"

	"github.com/mifwar/LinkSavvyBE/auth"
	"github.com/mifwar/LinkSavvyBE/db"
	"github.com/mifwar/LinkSavvyBE/db/models"
)

var Store = &session.Store{}

func init() {
	Store = session.New()
}

func Tags(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	tags := db.GetTags(userID.(float64))

	jsonTags, err := json.Marshal(tags)
	if err != nil {
		return err
	}

	return c.SendString(string(jsonTags))
}

func Categories(c *fiber.Ctx) error {
	userID := c.Locals("userID")
	categories := db.GetCategories(userID.(float64))

	jsonCategories, err := json.Marshal(categories)
	if err != nil {
		return err
	}

	return c.SendString(string(jsonCategories))
}

func Tag(c *fiber.Ctx) error {

	tag := models.Entity{}

	userID := c.Locals("userID")

	if userID == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error at accessing local context",
		})
	}

	if err := c.BodyParser(&tag); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "failed to parse",
		})
	}

	if err := db.CreateTag(tag, userID.(float64)); err != nil {
		fmt.Println("err: ", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "failed to insert the tag to DB",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "accepted"},
	)
}

func Category(c *fiber.Ctx) error {

	category := models.Entity{}

	userID := c.Locals("userID")

	if userID == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error at accessing local context",
		})
	}

	if err := c.BodyParser(&category); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "failed to parse",
		})
	}

	if err := db.CreateCategory(category, userID.(float64)); err != nil {
		fmt.Println("err: ", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "failed to insert the category to DB",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "accepted"},
	)
}

func Url(c *fiber.Ctx) error {
	return nil
}

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
		// return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		// 	"message": "invalid state",
		// })
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
