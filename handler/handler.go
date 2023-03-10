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
	"github.com/golang-jwt/jwt"

	"github.com/google/uuid"

	"github.com/mifwar/LinkSavvyBE/auth"
	"github.com/mifwar/LinkSavvyBE/db"
	"github.com/mifwar/LinkSavvyBE/db/models"
)

var Store = &session.Store{}

func init() {
	fmt.Println("init handler")
	Store = session.New()
	Store.RegisterType("string")
	fmt.Println("ok")
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

	authHeader := c.Get("Authorization")

	fmt.Println("authHeader: ", authHeader)

	authKey := strings.Replace(authHeader, "Bearer ", "", -1)

	token, err := jwt.Parse(authKey, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("JWT_KEY")), nil
	})

	if err != nil {
		fmt.Println("error jwt parse")
		return Logout(c)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message ": "unauthorized",
		})
	}

	if err != nil {
		fmt.Println("error string to int conversion")
		log.Fatalln(err)
	}

	userEmail := db.GetByUserID("email", claims["user"].(float64))
	name := db.GetByUserID("full_name", claims["user"].(float64))

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"email": userEmail,
		"name":  name,
	})
}

func Login(c *fiber.Ctx) error {

	user := models.User{}

	if err := c.BodyParser(&user); err != nil {
		return err
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
		return err
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

	fmt.Println("state set: ", state)
	session.Set("state", state)
	sessionBeforeSave := session.Get("state")
	fmt.Println("sessionBeforeSave: ", sessionBeforeSave)

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
