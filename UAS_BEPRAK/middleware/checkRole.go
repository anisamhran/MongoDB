package middleware

import (
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Middleware untuk memeriksa role user berdasarkan role_id
func CheckRole(requiredRoleID primitive.ObjectID) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Ambil role_id dari context yang disimpan setelah validasi JWT
		userRoleID, ok := c.Locals("role_id").(primitive.ObjectID)
		if !ok || userRoleID != requiredRoleID {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied. Insufficient role.",
			})
		}
		return c.Next()
	}
}
