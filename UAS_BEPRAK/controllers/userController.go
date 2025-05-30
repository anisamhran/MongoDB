package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"project-crud/config"
	"project-crud/middleware"
	"project-crud/models"
)

var userCollection *mongo.Collection = config.GetCollection("users")

// Fungsi login untuk memverifikasi user dan memberikan token
func Login(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    var inputUser models.User
    if err := c.BodyParser(&inputUser); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
    }

    // Cari user berdasarkan username di database
    var user models.User
    err := userCollection.FindOne(ctx, bson.M{"username": inputUser.Username}).Decode(&user)
    if err != nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
    }

    // Verifikasi password menggunakan bcrypt
    err = bcrypt.CompareHashAndPassword([]byte(user.Pass), []byte(inputUser.Pass))
    if err != nil {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid password"})
    }

    // Middleware Auth - Verifikasi Token JWT
    token, err := middleware.GenerateToken(user.Username, user.RoleID, user.JenisUserID)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate token"})
    }

    // Middleware CheckRole - Verifikasi apakah role user sesuai
    if user.RoleID == primitive.NilObjectID {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User has no assigned role"})
    }

    var role models.Role
    // Pastikan Anda mencari role dari roleCollection, bukan userCollection
    err = roleCollection.FindOne(ctx, bson.M{"_id": user.RoleID}).Decode(&role)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve role"})
    }

    // Middleware CheckJenisUser - Verifikasi apakah jenis user valid
    if user.JenisUserID == primitive.NilObjectID {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User has no assigned jenis user"})
    }

    var jenisUser models.JenisUser
    err = jenisUserCollection.FindOne(ctx, bson.M{"_id": user.JenisUserID}).Decode(&jenisUser)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve jenis user"})
    }

    // Jika semua validasi berhasil, kirimkan token
    return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "token":      token,
        "role":       role.Name,
        "jenis_user": jenisUser.NmJenisUser,
    })
}


// CreateUser membuat user baru
func CreateUser(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    var user models.User
    if err := c.BodyParser(&user); err != nil {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
    }

    // Ambil username dari user yang sedang login
    loggedInUsername, ok := c.Locals("username").(string)
    if !ok || loggedInUsername == "" {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized access"})
    }

    // Periksa apakah username sudah ada
    var existingUser models.User
    err := userCollection.FindOne(ctx, bson.M{"username": user.Username}).Decode(&existingUser)
    if err == nil {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Username already exists"})
    }
    if err != mongo.ErrNoDocuments {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }

    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Pass), bcrypt.DefaultCost)
    if err != nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash password"})
    }
    user.Pass = string(hashedPassword)

    // Generate ID untuk user baru
    user.ID = primitive.NewObjectID()

    // Waktu pembuatan
    loc, _ := time.LoadLocation("Asia/Jakarta")
    now := primitive.NewDateTimeFromTime(time.Now().In(loc))
    user.CreatedAt = now
    user.UpdatedAt = now

    // Isi CreatedBy dan UpdatedBy dengan logged-in username
    user.CreatedBy = loggedInUsername
    user.UpdatedBy = loggedInUsername

    // Ambil jenis user
    var jenisUser models.JenisUser
    err = jenisUserCollection.FindOne(ctx, bson.M{"_id": user.JenisUserID}).Decode(&jenisUser)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid jenis user ID"})
        }
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }

    // Isi user_modul dari template_modul di jenis_user
    if len(jenisUser.TemplateModul) == 0 {
        user.UserModul = []models.UserModul{}
    } else {
        for _, tmpl := range jenisUser.TemplateModul {
            var modul models.Modul
            err := modulCollection.FindOne(ctx, bson.M{"_id": tmpl.ModulID}).Decode(&modul)
            if err != nil {
                return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Modul not found"})
            }

            user.UserModul = append(user.UserModul, models.UserModul{
                ModulID:   tmpl.ModulID,
                NamaModul: modul.Name,
                CreatedAt: now,
                CreatedBy: loggedInUsername,
                UpdatedAt: now,
                UpdatedBy: loggedInUsername,
            })
        }
    }

    // Simpan user ke database
    _, err = userCollection.InsertOne(ctx, user)
    if err != nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }

    return c.Status(http.StatusCreated).JSON(user)
}


// GetAllUsers mendapatkan semua user dari koleksi
func GetAllUsers(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Menyimpan semua user dalam slice
    var users []models.User

    // Query untuk mendapatkan semua data user
    cursor, err := userCollection.Find(ctx, bson.M{})
    if err != nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }
    defer cursor.Close(ctx)

    for cursor.Next(ctx) {
        var user models.User
        if err := cursor.Decode(&user); err != nil {
            return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
        }
        users = append(users, user)
    }

    return c.Status(http.StatusOK).JSON(users)
}


// GetUserByID mendapatkan user berdasarkan ID
func GetUserByID(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Ambil ID dari URL parameter
    id := c.Params("id")
    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID format"})
    }

    // Cari user berdasarkan ID
    var user models.User
    err = userCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
        }
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }

    return c.Status(http.StatusOK).JSON(user)
}


// EditUser memperbarui data user berdasarkan ID
func EditUser(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Ambil ID dari URL parameter
    id := c.Params("id")
    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID format"})
    }

    // Parse input dari body request
    var updateData map[string]interface{}
    if err := c.BodyParser(&updateData); err != nil {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
    }

    // Hindari perubahan pada _id
    delete(updateData, "_id")

    // Tambahkan updated_at
    loc, _ := time.LoadLocation("Asia/Jakarta")
    updateData["updated_at"] = primitive.NewDateTimeFromTime(time.Now().In(loc))

    // Update user berdasarkan ID
    update := bson.M{"$set": updateData}
    result, err := userCollection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
    if err != nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }

    if result.MatchedCount == 0 {
        return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
    }

    return c.Status(http.StatusOK).JSON(fiber.Map{"message": "User updated successfully"})
}

// EditJenisUserFromUser mengubah jenis user pada user berdasarkan ID
func EditJenisUserFromUser(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Ambil ID dari URL parameter
    id := c.Params("id")
    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID format"})
    }

    // Parse input dari body request
    var request struct {
        JenisUserID string `json:"jenis_user_id"`
    }

    if err := c.BodyParser(&request); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
    }

    // Validasi jenis_user_id
    jenisUserID, err := primitive.ObjectIDFromHex(request.JenisUserID)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid jenis_user_id format"})
    }

    // Ambil data jenis user berdasarkan jenis_user_id
    var jenisUser models.JenisUser
    err = jenisUserCollection.FindOne(ctx, bson.M{"_id": jenisUserID}).Decode(&jenisUser)
    if err != nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Jenis user not found"})
    }

    // Waktu sekarang
    loc, _ := time.LoadLocation("Asia/Jakarta")
    now := primitive.NewDateTimeFromTime(time.Now().In(loc))

    // Filter untuk user yang akan diperbarui
    filter := bson.M{"_id": objectID}

    // Mulai membuat update document
    updateFields := bson.M{
        "$set": bson.M{
            "jenis_user_id": jenisUserID,  // Memperbarui jenis_user_id
            "updated_at": now,             // Perbarui waktu
            "user_modul": jenisUser.TemplateModul, // Set user_modul (kosong jika tidak ada)
        },
    }

    // Jika TemplateModul kosong, set user_modul sebagai array kosong
    if len(jenisUser.TemplateModul) == 0 {
        updateFields["$set"].(bson.M)["user_modul"] = []interface{}{}
    }

    // Lakukan update user
    result, err := userCollection.UpdateOne(ctx, filter, updateFields)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update user type"})
    }

    // Jika tidak ada data yang diupdate
    if result.MatchedCount == 0 {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
    }

    return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "message":   "User type updated successfully",
        "user_id":   objectID,
        "jenis_user_id": request.JenisUserID,
        "user_modul": jenisUser.TemplateModul,  // Menyertakan modul yang baru
    })
}


// DeleteUser menghapus user berdasarkan ID
func DeleteUser(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Ambil ID dari URL parameter
    id := c.Params("id")
    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ID format"})
    }

    // Hapus user berdasarkan ID
    result, err := userCollection.DeleteOne(ctx, bson.M{"_id": objectID})
    if err != nil {
        return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
    }

    if result.DeletedCount == 0 {
        return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
    }

    return c.Status(http.StatusOK).JSON(fiber.Map{"message": "User deleted successfully"})
}


// Menambahkan Tambahan Modul Baru
func AddUserModule(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Ambil input dari body request
    var request struct {
        UserID    primitive.ObjectID `json:"user_id"`
        ModulID   primitive.ObjectID `json:"modul_id"`
    }

    if err := c.BodyParser(&request); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
    }

    // Validasi ObjectID
    if !primitive.IsValidObjectID(request.UserID.Hex()) || !primitive.IsValidObjectID(request.ModulID.Hex()) {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid ObjectID format"})
    }

    // Ambil informasi modul dari koleksi Modul
    var modul models.Modul
    err := modulCollection.FindOne(ctx, bson.M{"_id": request.ModulID}).Decode(&modul)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Modul not found"})
        }
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error fetching modul", "details": err.Error()})
    }

    // Persiapkan data modul baru
    loc, _ := time.LoadLocation("Asia/Jakarta")
    now := primitive.NewDateTimeFromTime(time.Now().In(loc))

    newModule := models.UserModul{
        ModulID:    modul.ID,
        NamaModul:  modul.Name,
        CreatedAt:  now,
        UpdatedAt:  now,
    }

    // Tambahkan modul ke array user_modul
    update := bson.M{
        "$addToSet": bson.M{"user_modul": newModule}, // Hindari duplikasi
    }

    // Lakukan update pada userCollection
    _, err = userCollection.UpdateByID(ctx, request.UserID, update)
    if err != nil {
        fmt.Println("MongoDB Update Error:", err) // Debugging
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to add module", "details": err.Error()})
    }

    return c.JSON(fiber.Map{
        "message": "Module added successfully",
        "modul":   newModule,
    })
}


// Menghapus Modul Tambahan
func RemoveUserModule(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Ambil input dari body request
	var request struct {
		UserID  primitive.ObjectID `json:"user_id"`  // ID User
		ModulID primitive.ObjectID `json:"modul_id"` // ID Modul yang akan dihapus
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Update query untuk menghapus modul dari array
	update := bson.M{
		"$pull": bson.M{"user_modul": bson.M{"modul_id": request.ModulID}}, // Hapus berdasarkan modul_id
	}

	// Lakukan update
	_, err := userCollection.UpdateByID(ctx, request.UserID, update)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to remove module"})
	}

	return c.JSON(fiber.Map{
		"message": "Module removed successfully",
		"modul_id": request.ModulID,
	})
}


// Memperbarui Modul Tambahan
func UpdateUserModule(c *fiber.Ctx) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Ambil input dari body request
    var request struct {
        UserID    primitive.ObjectID `json:"user_id"`    // ID User
        ModulID   primitive.ObjectID `json:"modul_id"`   // ID Modul yang akan diupdate
        UpdatedBy string             `json:"updated_by"` // User yang memperbarui
        UpdateIP  string             `json:"update_ip"`  // IP yang digunakan
    }

    if err := c.BodyParser(&request); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
    }

    // Waktu sekarang
    loc, _ := time.LoadLocation("Asia/Jakarta")
    now := primitive.NewDateTimeFromTime(time.Now().In(loc))

    // Update spesifik elemen dalam array UserModul
    filter := bson.M{
        "_id":             request.UserID,                      // Filter User ID
        "user_modul.modul_id": request.ModulID,                 // Cari elemen dalam array
    }

    update := bson.M{
        "$set": bson.M{
            "user_modul.$.updated_at": now,          // Update waktu terakhir diperbarui
            "user_modul.$.updated_by": request.UpdatedBy,
            "user_modul.$.update_ip":  request.UpdateIP,
        },
    }

    // Lakukan update
    _, err := userCollection.UpdateOne(ctx, filter, update)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update module"})
    }

    return c.JSON(fiber.Map{
        "message": "Module updated successfully",
        "modul_id": request.ModulID,
    })
}

