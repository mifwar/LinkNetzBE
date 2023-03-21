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

func GetTags(userID float64) []models.Entity {
	sqlStatement := "SELECT name, emoji FROM tags WHERE user_id = $1"
	rows, err := db.Query(sqlStatement, userID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var tags []models.Entity

	for rows.Next() {
		var tag models.Entity
		err := rows.Scan(&tag.Name, &tag.Emoji)
		if err != nil {
			log.Fatal(err)
		}
		tags = append(tags, tag)
	}

	return tags
}

func CreateTag(newTags models.Entity, userID float64) error {
	sqlStatement := "INSERT INTO tags (name, emoji, user_id) VALUES ($1, $2, $3)"
	_, err := db.Exec(sqlStatement, newTags.Name, newTags.Emoji, userID)

	if err != nil {
		return err
	}

	return nil
}

func GetCategories(userID float64) []models.Entity {
	sqlStatement := "SELECT name, emoji FROM categories WHERE user_id = $1"
	rows, err := db.Query(sqlStatement, userID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var categories []models.Entity

	for rows.Next() {
		var category models.Entity
		err := rows.Scan(&category.Name, &category.Emoji)
		if err != nil {
			log.Fatal(err)
		}
		categories = append(categories, category)
	}

	return categories
}

func CreateCategory(newCategory models.Entity, userID float64) error {
	sqlStatement := "INSERT INTO categories (name, emoji, user_id) VALUES ($1, $2, $3)"
	_, err := db.Exec(sqlStatement, newCategory.Name, newCategory.Emoji, userID)

	if err != nil {
		return err
	}

	return nil
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
