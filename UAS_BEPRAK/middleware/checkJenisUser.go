package middleware

import (
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Middleware untuk memeriksa jenis user berdasarkan jenis_user_id
func CheckJenisUser(requiredJenisUserID primitive.ObjectID) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Ambil jenis_user_id dari context yang disimpan setelah validasi JWT
		userJenisUserID, ok := c.Locals("jenis_user_id").(primitive.ObjectID)
		if !ok || userJenisUserID != requiredJenisUserID {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied. Incorrect jenis user.",
			})
		}
		return c.Next()
	}
}
