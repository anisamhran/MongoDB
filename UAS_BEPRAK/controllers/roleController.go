package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"project-crud/config"
	"project-crud/models"
)

var roleCollection *mongo.Collection = config.GetCollection("roles")

//  Create Role
func CreateRole(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    var role models.Role
    if err := c.BodyParser(&role); err != nil {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
    }

    // Ambil username dari user yang sedang login
    loggedInUsername, ok := c.Locals("username").(string)
    if !ok || loggedInUsername == "" {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized access"})
    }

    // Validasi input
    if role.Name == "" {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Role name is required"})
    }

    // Ambil username dari context (middleware)
    username := c.Locals("username").(string)
    if username == "" {
        return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "User not authenticated"})
    }

    // Cek apakah role dengan nama yang sama sudah ada
    var existingRole models.Role
    err := roleCollection.FindOne(ctx, bson.M{"name": role.Name}).Decode(&existingRole)
    if err == nil {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Role name already exists"})
    }

    if err != mongo.ErrNoDocuments {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }

    // Set waktu dan user yang membuat
    loc, _ := time.LoadLocation("Asia/Jakarta")
    role.ID = primitive.NewObjectID()
    role.CreatedAt = primitive.NewDateTimeFromTime(time.Now().In(loc))
    role.UpdatedAt = role.CreatedAt

     // Isi CreatedBy dan UpdatedBy dengan logged-in username
     role.CreatedBy = loggedInUsername
     role.UpdatedBy = loggedInUsername

    // Masukkan role ke database
    _, err = roleCollection.InsertOne(ctx, role)
    if err != nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }

    return c.Status(http.StatusCreated).JSON(role)
}


// GetRoles untuk mendapatkan semua role
func GetRoles(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Query semua role dari koleksi
    cursor, err := roleCollection.Find(ctx, bson.M{})
    if err != nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch roles"})
    }
    defer cursor.Close(ctx)

    // Iterasi hasil query
    var roles []models.Role
    for cursor.Next(ctx) {
        var role models.Role
        if err := cursor.Decode(&role); err != nil {
            return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to decode role"})
        }
        roles = append(roles, role)
    }

    if err := cursor.Err(); err != nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Cursor error: " + err.Error()})
    }

    // Kembalikan daftar role
    return c.Status(http.StatusOK).JSON(roles)
}


// GetRoles by ID
func GetRole(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    roleID := c.Params("id")
    objID, err := primitive.ObjectIDFromHex(roleID)
    if err != nil {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid role ID"})
    }

    var role models.Role
    err = roleCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&role)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Role not found"})
        }
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }

    return c.Status(http.StatusOK).JSON(role)
}


// Update Role by ID
func EditRole(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Ambil ID dari parameter
    roleID := c.Params("id")
    objID, err := primitive.ObjectIDFromHex(roleID)
    if err != nil {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid role ID"})
    }

    // Parsing input
    var roleInput models.Role
    if err := c.BodyParser(&roleInput); err != nil {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
    }

    // Validasi input
    if roleInput.Name == "" {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Role name is required"})
    }

    // Ambil username dari context (middleware)
    username := c.Locals("username").(string)
    if username == "" {
        return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "User not authenticated"})
    }

    // Update data
    loc, _ := time.LoadLocation("Asia/Jakarta")
    update := bson.M{
        "$set": bson.M{
            "name":       roleInput.Name,
            "updated_at": primitive.NewDateTimeFromTime(time.Now().In(loc)),
            "updated_by": username,
        },
    }

    result, err := roleCollection.UpdateOne(ctx, bson.M{"_id": objID}, update)
    if err != nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }

    if result.MatchedCount == 0 {
        return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Role not found"})
    }

    return c.Status(http.StatusOK).JSON(fiber.Map{"message": "Role updated successfully"})
}


// Delete Role
func DeleteRole(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Ambil ID dari parameter
    roleID := c.Params("id")
    objID, err := primitive.ObjectIDFromHex(roleID)
    if err != nil {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid role ID"})
    }

    // Hapus data
    result, err := roleCollection.DeleteOne(ctx, bson.M{"_id": objID})
    if err != nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }

    if result.DeletedCount == 0 {
        return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Role not found"})
    }

    return c.Status(http.StatusOK).JSON(fiber.Map{"message": "Role deleted successfully"})
}
