package main

import (
	"log"
	"project-crud/routes"

	"github.com/gofiber/fiber/v2"
)


func main() {
    app := fiber.New()

    routes.RouterApp(app)

    log.Fatal(app.Listen(":3000"))
}