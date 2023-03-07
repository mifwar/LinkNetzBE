package auth

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/mifwar/LinkSavvyBE/db"
	"golang.org/x/crypto/bcrypt"
)

func GenerateJWT(userID int) string {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Add(24 * time.Hour).Unix()
	claims["authorized"] = true
	claims["user"] = userID

	secretKey := os.Getenv("JWT_KEY")
	key := []byte(secretKey)

	tokenString, err := token.SignedString(key)

	if err != nil {
		log.Fatal("can't sign JWT", err)
	}

	return tokenString
}

func ValidateEmail(email string) error {
	storedEmail := db.GetEmail(email)
	if email != storedEmail {
		return errors.New("Email is not registered")
	}
	return nil
}

func ValidatePassword(email string, password string) error {

	hashedPassword := db.GetPassword(email)

	if err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(password)); err != nil {
		return err
	}
	return nil
}

func ValidateLoginMethod(email, loginMethod string) error {

	storedLoginMethod := db.GetLoginMethod(email)

	if loginMethod != storedLoginMethod {
		var errorMessage string

		if storedLoginMethod == "email" {
			errorMessage = "This email is already registered using email and password. Please sign using your email and password"
		} else {
			errorMessage = "This email already registered using Google Authentication. Please use Google to sign in using this email"
		}

		return errors.New(errorMessage)
	}

	return nil
}
