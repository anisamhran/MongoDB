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

var kategoriModulCollection *mongo.Collection = config.GetCollection("kategori_modul")// Pastikan ini sudah diinisialisasi di main.go

// CreateKategoriModul membuat kategori modul baru
func CreateKategoriModul(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

     // Ambil username dari user yang sedang login
     loggedInUsername, ok := c.Locals("username").(string)
     if !ok || loggedInUsername == "" {
         return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized access"})
     }

    var kategoriModul models.KategoriModul
    if err := c.BodyParser(&kategoriModul); err != nil {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
    }

    // Validasi input
    if kategoriModul.Name == "" {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Name is required"})
    }

    // Set data tambahan
    loc, err := time.LoadLocation("Asia/Jakarta")
    if err != nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load time location"})
    }
    kategoriModul.ID = primitive.NewObjectID()
    // Isi CreatedBy dan UpdatedBy dengan logged-in username
    kategoriModul.CreatedBy = loggedInUsername
    kategoriModul.UpdatedBy = loggedInUsername
    kategoriModul.CreatedAt = primitive.NewDateTimeFromTime(time.Now().In(loc))
    kategoriModul.UpdatedAt = kategoriModul.CreatedAt

    // Insert ke database
    _, err = kategoriModulCollection.InsertOne(ctx, kategoriModul)
    if err != nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create kategori modul"})
    }

    return c.Status(http.StatusCreated).JSON(kategoriModul)
}


//  Get semua kategori modul
func GetAllKategoriModul(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Query semua kategori modul dari koleksi
    cursor, err := kategoriModulCollection.Find(ctx, bson.M{})
    if err != nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch kategori modul"})
    }
    defer cursor.Close(ctx)

    // Iterasi hasil query
    var kategoriModul []models.KategoriModul
    for cursor.Next(ctx) {
        var modul models.KategoriModul
        if err := cursor.Decode(&modul); err != nil {
            return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to decode kategori modul"})
        }
        kategoriModul = append(kategoriModul, modul)
    }

    if err := cursor.Err(); err != nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Cursor error: " + err.Error()})
    }

    // Kembalikan daftar kategori modul
    return c.Status(http.StatusOK).JSON(kategoriModul)
}


//  Get kategori modul by ID
func GetKategoriModulByID(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Ambil ID dari parameter
    kategoriID := c.Params("id")
    objID, err := primitive.ObjectIDFromHex(kategoriID)
    if err != nil {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid kategori modul ID"})
    }

    // Cari kategori modul berdasarkan ID
    var kategoriModul models.KategoriModul
    err = kategoriModulCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&kategoriModul)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Kategori modul not found"})
        }
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }

    // Kembalikan kategori modul
    return c.Status(http.StatusOK).JSON(kategoriModul)
}


// Edit Kategori by ID
func EditKategoriModul(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Ambil ID dari parameter
    kategoriID := c.Params("id")
    objID, err := primitive.ObjectIDFromHex(kategoriID)
    if err != nil {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid kategori modul ID"})
    }

    // Parsing input
    var kategoriInput models.KategoriModul
    if err := c.BodyParser(&kategoriInput); err != nil {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
    }

    // Validasi input
    if kategoriInput.Name == "" {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Name is required"})
    }

    // Ambil username dari middleware
    updatedBy := c.Locals("username").(string)

    // Update data
    loc, err := time.LoadLocation("Asia/Jakarta")
    if err != nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }

    update := bson.M{
        "$set": bson.M{
            "name":       kategoriInput.Name,
            "updated_at": primitive.NewDateTimeFromTime(time.Now().In(loc)),
            "updated_by": updatedBy,
        },
    }

    result, err := kategoriModulCollection.UpdateOne(ctx, bson.M{"_id": objID}, update)
    if err != nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }

    if result.MatchedCount == 0 {
        return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Kategori modul not found"})
    }

    return c.Status(http.StatusOK).JSON(fiber.Map{"message": "Kategori modul updated successfully"})
}


// Delete Kategori Modul by ID
func DeleteKategoriModul(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Ambil ID dari parameter
    kategoriID := c.Params("id")
    objID, err := primitive.ObjectIDFromHex(kategoriID)
    if err != nil {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid kategori modul ID"})
    }

    // Hapus kategori modul dari koleksi
    result, err := kategoriModulCollection.DeleteOne(ctx, bson.M{"_id": objID})
    if err != nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }

    if result.DeletedCount == 0 {
        return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Kategori modul not found"})
    }

    return c.Status(http.StatusOK).JSON(fiber.Map{"message": "Kategori modul deleted successfully"})
}
