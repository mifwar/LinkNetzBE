package handler

import (
	"github.com/gofiber/fiber/v2"
)

func CreateCategory(c *fiber.Ctx) error {
	return CreateEntity(c, "categories")
}

func GetCategories(c *fiber.Ctx) error {
	return GetEntities(c, "categories")
}

func EditCategory(c *fiber.Ctx) error {
	return EditEntity(c, "categories")
}

func DeleteCategory(c *fiber.Ctx) error {
	return DeleteEntity(c, "categories")
}
