package handler

import (
	"github.com/gofiber/fiber/v2"
)

func CreateTag(c *fiber.Ctx) error {
	return CreateEntity(c, "tags")
}

func GetTags(c *fiber.Ctx) error {
	return GetEntities(c, "tags")
}

func EditTag(c *fiber.Ctx) error {
	return EditEntity(c, "tags")
}

func DeleteTag(c *fiber.Ctx) error {
	return DeleteEntity(c, "tags")
}
