package controllers

import (
	"context"
	"net/http"
	"project-crud/config"
	"project-crud/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var jenisUserCollection *mongo.Collection = config.GetCollection("jenis_users")

// CreateJenisUser untuk membuat jenis user baru dengan template modul
func CreateJenisUser(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Ambil username dari user yang sedang login
    loggedInUsername, ok := c.Locals("username").(string)
    if !ok || loggedInUsername == "" {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized access"})
    }

    var jenisUser models.JenisUser
    if err := c.BodyParser(&jenisUser); err != nil {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
    }

    // Set data tambahan untuk jenis user
    jenisUser.ID = primitive.NewObjectID()
    loc, _ := time.LoadLocation("Asia/Jakarta")
    now := primitive.NewDateTimeFromTime(time.Now().In(loc))
    jenisUser.CreatedAt = now
    jenisUser.UpdatedAt = now
     // Isi CreatedBy dan UpdatedBy dengan logged-in username
     jenisUser.CreatedBy = loggedInUsername
     jenisUser.UpdatedBy = loggedInUsername

    // Validasi dan simpan modul
    for _, templateModul := range jenisUser.TemplateModul {
        if templateModul.ModulID.IsZero() {
            return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid Modul ID in template_modul"})
        }
    }

    // Simpan ke database
    _, err := jenisUserCollection.InsertOne(ctx, jenisUser)
    if err != nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }

    return c.Status(http.StatusCreated).JSON(jenisUser)
}


// GetAllJenisUser untuk mendapatkan semua jenis user
func GetAllJenisUser(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Ambil semua jenis user dari koleksi
    var jenisUsers []models.JenisUser
    cursor, err := jenisUserCollection.Find(ctx, bson.M{})
    if err != nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }
    defer cursor.Close(ctx)

    // Decode hasil query ke dalam slice jenisUsers
    for cursor.Next(ctx) {
        var jenisUser models.JenisUser
        if err := cursor.Decode(&jenisUser); err != nil {
            return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
        }
        jenisUsers = append(jenisUsers, jenisUser)
    }

    if err := cursor.Err(); err != nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }

    // Kembalikan response
    return c.Status(http.StatusOK).JSON(jenisUsers)
}


// GetJenisUserByID untuk mendapatkan jenis user berdasarkan ID
func GetJenisUserByID(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Ambil ID dari URL parameter
    id := c.Params("id")
    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID format"})
    }

    // Cari jenis user berdasarkan ID
    var jenisUser models.JenisUser
    err = jenisUserCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&jenisUser)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Jenis user not found"})
        }
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }

    // Kembalikan response
    return c.Status(http.StatusOK).JSON(jenisUser)
}


// EditJenisUser untuk mengedit jenis user berdasarkan ID
func EditJenisUser(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Ambil ID dari parameter
    id := c.Params("id")
    objID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID format"})
    }

    // Parse input dari request body
    var input models.JenisUser
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

    // Update hanya field yang diisi
    if input.NmJenisUser != "" {
        updateFields["nm_jenis_user"] = input.NmJenisUser
    }

    // Query update utama
    updateQuery := bson.M{"$set": updateFields}

    // Jika ada template_modul, tambahkan tanpa menghapus yang lama
    if len(input.TemplateModul) > 0 {
        for _, modul := range input.TemplateModul {
            if modul.ModulID.IsZero() || !primitive.IsValidObjectID(modul.ModulID.Hex()) {
                return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid modul_id format"})
            }
        }

        updateQuery["$addToSet"] = bson.M{
            "template_modul": bson.M{"$each": input.TemplateModul},
        }
    }

    // Update data di database
    result, err := jenisUserCollection.UpdateOne(ctx, bson.M{"_id": objID}, updateQuery)
    if err != nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }

    if result.MatchedCount == 0 {
        return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Jenis user not found"})
    }

    // Ambil data terbaru setelah update
    var updatedJenisUser models.JenisUser
    err = jenisUserCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&updatedJenisUser)
    if err != nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch updated data"})
    }

    // Kembalikan response
    return c.Status(http.StatusOK).JSON(updatedJenisUser)
}


// DeleteTemplateModul untuk menguraangi modul tertentu dari template_modul pada jenis user
func DeleteTemplateModul(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Ambil ID dari parameter
    id := c.Params("id")
    objID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID format"})
    }

    // Parse request body
    var input struct {
        ModulID string `json:"modul_id"`
    }
    if err := c.BodyParser(&input); err != nil {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
    }

    // Validasi modul_id
    if input.ModulID == "" || !primitive.IsValidObjectID(input.ModulID) {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid modul_id format"})
    }
    modulObjID, _ := primitive.ObjectIDFromHex(input.ModulID)

    // Query update untuk menghapus modul dari template_modul
    updateQuery := bson.M{
        "$pull": bson.M{
            "template_modul": bson.M{"modul_id": modulObjID},
        },
    }

    // Lakukan update di database
    result, err := jenisUserCollection.UpdateOne(ctx, bson.M{"_id": objID}, updateQuery)
    if err != nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }

    if result.MatchedCount == 0 {
        return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Jenis user not found"})
    }
    if result.ModifiedCount == 0 {
        return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "Modul not found in template_modul"})
    }

    // Ambil data terbaru setelah update
    var updatedJenisUser models.JenisUser
    err = jenisUserCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&updatedJenisUser)
    if err != nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch updated data"})
    }

    // Kembalikan response
    return c.Status(http.StatusOK).JSON(fiber.Map{
        "message": "Modul removed successfully",
        "data":    updatedJenisUser,
    })
}


// DeleteJenisUser untuk menghapus jenis user berdasarkan ID
func DeleteJenisUser(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Ambil ID dari URL parameter
    id := c.Params("id")
    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID format"})
    }

    // Hapus jenis user berdasarkan ID
    _, err = jenisUserCollection.DeleteOne(ctx, bson.M{"_id": objectID})
    if err != nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }

    // Kembalikan response
    return c.Status(http.StatusOK).JSON(fiber.Map{"message": "Jenis user deleted successfully"})
}
