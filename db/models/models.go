package models

import (
	"log"

	util "github.com/mifwar/LinkSavvyBE/utils"
)

type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type NewUser struct {
	FullName string `json:"fullname"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (u *NewUser) Encrypt() {
	hashedPassword, err := util.HashPassword(u.Password)

	if err != nil {
		log.Fatal(err)
	}

	u.Password = hashedPassword
}
