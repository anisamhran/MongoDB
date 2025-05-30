package controllers

import "github.com/gofiber/fiber/v2"

func HomeFunc(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message" : "Hello World Testing Baru Ni Bro",
	})
}