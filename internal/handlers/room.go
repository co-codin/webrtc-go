package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	guuid "github.com/google/uuid"
)

func RoomCreate(c fiber.Ctx) error {
	roomID := guuid.New().String()
	url := fmt.Sprintf("/room/%s", roomID)
	return c.Redirect().To(url)
}

func Room(c fiber.Ctx) error {
	return c.SendString("Room")
}
