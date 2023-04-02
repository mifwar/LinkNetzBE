package handler

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"

	"github.com/mifwar/LinkSavvyBE/db"
	"github.com/mifwar/LinkSavvyBE/db/models"
)

var Store = &session.Store{}

func init() {
	Store = session.New()
}

func DeleteEntity(c *fiber.Ctx, entityType string) error {
	userID := c.Locals("userID")

	if userID == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error at accessing local context",
		})
	}

	entityID, err := strconv.Atoi(c.Params("id"))

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error at converting entityID",
		})
	}

	switch entityType {
	case "tags":
		if err := db.DeleteTags(entityID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "error at deleting tag",
			})
		}
	case "categories":
		if err := db.DeleteCategory(entityID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "error at deleting category",
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": fmt.Sprintf("%s successfully deleted", entityType)},
	)
}

func EditEntity(c *fiber.Ctx, entityType string) error {
	userID := c.Locals("userID")

	if userID == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error at accessing local context",
		})
	}

	var newEntity models.Entity

	if err := c.BodyParser(&newEntity); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error at parsing body",
		})
	}

	switch entityType {
	case "tags":
		if err := db.EditTags(newEntity); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "error at editing tag",
			})
		}
	case "categories":
		if err := db.EditCategory(newEntity); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "error at editing category",
			})
		}
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "invalid entity type",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": fmt.Sprintf("%s successfully edited", entityType)},
	)
}

func GetEntities(c *fiber.Ctx, entityType string) error {

	userID := c.Locals("userID")

	if userID == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error at accessing local context",
		})
	}

	entities := []models.Entity{}

	var dbEntities []models.Entity

	switch entityType {
	case "tags":
		dbEntities = db.GetTags(userID.(float64))
	case "categories":
		dbEntities = db.GetCategories(userID.(float64))
	default:
		return c.Status(fiber.StatusBadRequest).SendString("invalid entity type")
	}

	if len(dbEntities) > 0 {
		entities = dbEntities
	}

	jsonEntities, err := json.Marshal(entities)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error at marshalling entities",
		})
	}

	return c.SendString(string(jsonEntities))
}

func CreateEntity(c *fiber.Ctx, entityType string) error {
	userID := c.Locals("userID")

	if userID == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error at accessing local context",
		})
	}

	var newEntity models.Entity

	if err := c.BodyParser(&newEntity); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "error at parsing body",
		})
	}

	switch entityType {
	case "tags":
		if err := db.CreateTag(newEntity, userID.(float64)); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "error at creating tag",
			})
		}
	case "categories":
		if err := db.CreateCategory(newEntity, userID.(float64)); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "error at creating category",
			})
		}
	default:
		return c.Status(fiber.StatusBadRequest).SendString("invalid entity type")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": fmt.Sprintf("%s successfully created", entityType)},
	)
}

func Url(c *fiber.Ctx) error {
	return nil
}
