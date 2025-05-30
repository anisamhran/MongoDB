package controllers

import (
	"context"
	"net/http"
	"time"

	"project-crud/config" // Ganti dengan nama modul Anda
	"project-crud/models" // Ganti dengan nama modul Anda

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var modulCollection *mongo.Collection = config.GetCollection("moduls") 

// CreateModul membuat modul baru
func CreateModul(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Ambil username dari user yang sedang login
    loggedInUsername, ok := c.Locals("username").(string)
    if !ok || loggedInUsername == "" {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized access"})
    }

    // Parse input dari request body
    var modul models.Modul
    if err := c.BodyParser(&modul); err != nil {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
    }

    // Validasi input
    if modul.Name == "" {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Name is required"})
    }
    if modul.Description == "" {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Description is required"})
    }
    if modul.KategoriModul.IsZero() {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "KategoriModul ID is required"})
    }
    if modul.AlamatURL == "" {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "AlamatURL is required"})
    }
    if modul.GbrIcon == "" {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "GbrIcon is required"})
    }

    // Validasi keberadaan KategoriModul
    var kategoriModul models.KategoriModul
    err := kategoriModulCollection.FindOne(ctx, bson.M{"_id": modul.KategoriModul}).Decode(&kategoriModul)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid KategoriModul ID"})
        }
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }

    // Set data tambahan
    loc, err := time.LoadLocation("Asia/Jakarta")
    if err != nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load time location"})
    }
    modul.ID = primitive.NewObjectID()
    // Isi CreatedBy dan UpdatedBy dengan logged-in username
    modul.CreatedBy = loggedInUsername
    modul.UpdatedBy = loggedInUsername
    modul.CreatedAt = primitive.NewDateTimeFromTime(time.Now().In(loc))
    modul.UpdatedAt = modul.CreatedAt

    // Masukkan data ke database
    _, err = modulCollection.InsertOne(ctx, modul)
    if err != nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create modul"})
    }

    // Kembalikan response
    return c.Status(http.StatusCreated).JSON(modul)
}


// GetAllModul mendapatkan semua modul
func GetAllModul(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Ambil data dari database
    cursor, err := modulCollection.Find(ctx, bson.M{})
    if err != nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }
    var moduls []models.Modul
    if err := cursor.All(ctx, &moduls); err != nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }

    // Kembalikan response
    return c.Status(http.StatusOK).JSON(moduls)
}


// GetModulByID mendapatkan modul berdasarkan ID
func GetModulByID(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Ambil ID dari parameter
    id := c.Params("id")
    objID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID format"})
    }

    // Cari modul di database
    var modul models.Modul
    err = modulCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&modul)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Modul not found"})
        }
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }

    // Kembalikan response
    return c.Status(http.StatusOK).JSON(modul)
}


// EditModul memperbarui modul berdasarkan ID
func EditModul(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Ambil ID dari parameter
    id := c.Params("id")
    objID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID format"})
    }

    // Parse input dari request body
    var input models.Modul
    if err := c.BodyParser(&input); err != nil {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
    }

    // Ambil username dari middleware JWT
    updatedBy := c.Locals("username").(string)

    // Set data tambahan
    loc, err := time.LoadLocation("Asia/Jakarta")
    if err != nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to load time location"})
    }

    updateFields := bson.M{
        "updated_by": updatedBy,
        "updated_at": primitive.NewDateTimeFromTime(time.Now().In(loc)),
    }

    // Update hanya field yang tidak kosong
    if input.Name != "" {
        updateFields["name"] = input.Name
    }
    if input.Description != "" {
        updateFields["description"] = input.Description
    }
    if input.AlamatURL != "" {
        updateFields["alamat_url"] = input.AlamatURL
    }
    if input.GbrIcon != "" {
        updateFields["gbr_icon"] = input.GbrIcon
    }

    // Lakukan update
    update := bson.M{
        "$set": updateFields,
    }

    // Update data ke database
    result, err := modulCollection.UpdateOne(ctx, bson.M{"_id": objID}, update)
    if err != nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }

    if result.MatchedCount == 0 {
        return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Modul not found"})
    }

    // Ambil data terbaru setelah update
    var updatedModul models.Modul
    err = modulCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&updatedModul)
    if err != nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch updated data"})
    }

    // Kembalikan response
    return c.Status(http.StatusOK).JSON(updatedModul)
}


// DeleteModul menghapus modul berdasarkan ID
func DeleteModul(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Ambil ID dari parameter
    id := c.Params("id")
    objID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID format"})
    }

    // Hapus data modul
    _, err = modulCollection.DeleteOne(ctx, bson.M{"_id": objID})
    if err != nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }

    // Kembalikan response
    return c.Status(http.StatusOK).JSON(fiber.Map{"message": "Modul deleted successfully"})
}
