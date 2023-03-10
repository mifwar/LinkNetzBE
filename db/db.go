package db

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/mifwar/LinkSavvyBE/db/models"

	"database/sql"

	_ "github.com/lib/pq"
)

var db *sql.DB

func init() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("failed to access env")
	}

	dbUserName := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("PORT")

	dbCommand := fmt.Sprintf("postgres://%s:%s@localhost:%s/%s?sslmode=disable", dbUserName, dbPassword, dbPort, dbName)

	db, err = sql.Open("postgres", dbCommand)

	if err != nil {
		log.Fatal("failed to access db")
	}
}

func CreateUser(newUser models.NewUser, loginMethod string) error {

	sqlStatement := "INSERT INTO users (full_name, email, password, login_method) VALUES ($1, $2, $3, $4)"
	_, err := db.Exec(sqlStatement, newUser.FullName, newUser.Email, newUser.Password, loginMethod)

	if err != nil {
		return err
	}

	return nil
}

func GetByUserID(columnName string, userID float64) string {
	sqlStatement := fmt.Sprintf("SELECT %s FROM users WHERE id = $1", columnName)
	result := ""

	err := db.QueryRow(sqlStatement, userID).Scan(&result)

	if err != nil {
		return ""
	}

	return result
}

func GetByEmail(columnName, email string) string {
	sqlStatement := fmt.Sprintf("SELECT %s FROM users WHERE email = $1", columnName)
	result := ""

	err := db.QueryRow(sqlStatement, email).Scan(&result)

	if err != nil {
		return ""
	}

	return result
}

func GetEmail(email string) string {
	return GetByEmail("email", email)
}

func GetLoginMethod(email string) string {
	return GetByEmail("login_method", email)
}

func GetUserID(email string) int {
	result, err := strconv.Atoi(GetByEmail("id", email))

	if err != nil {
		log.Fatalln(err)
	}

	return result
}

func GetPassword(email string) []byte {
	return []byte(GetByEmail("password", email))
}
