package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Secret key untuk enkripsi
var SecretKey = []byte("anisa")

// Fungsi untuk membuat JWT
func GenerateToken(username string, roleID, jenisUserID primitive.ObjectID) (string, error) {
	// Header JWT
	header := map[string]string{
		"alg": "HS256", // Algoritma yang digunakan
		"typ": "JWT",   // Tipe token
	}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	headerEncoded := base64.URLEncoding.EncodeToString(headerJSON)

	// Payload JWT (Data yang ingin dikirimkan)
	payload := map[string]interface{}{
		"username":       username,
		"role_id":        roleID.Hex(),
		"jenis_user_id":  jenisUserID.Hex(),
		"exp":            time.Now().Add(time.Hour * 1).Unix(), // Token berlaku selama 1 jam
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	payloadEncoded := base64.URLEncoding.EncodeToString(payloadJSON)

	// Signature - menggunakan HMAC SHA256 dengan secret key
	message := headerEncoded + "." + payloadEncoded
	h := hmac.New(sha256.New, SecretKey)
	h.Write([]byte(message))
	signature := base64.URLEncoding.EncodeToString(h.Sum(nil))

	// Gabungkan semuanya menjadi token
	token := headerEncoded + "." + payloadEncoded + "." + signature
	return token, nil
}


// Fungsi untuk memverifikasi JWT
func JWTAuth(c *fiber.Ctx) error {
	// Ambil token dari header Authorization
	token := c.Get("Authorization")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "No token provided",
		})
	}

	// Periksa apakah token menggunakan format Bearer
	if !strings.HasPrefix(token, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token format",
		})
	}

	// Ambil token yang sebenarnya dengan menghapus "Bearer "
	token = strings.TrimPrefix(token, "Bearer ")

	// Pisahkan token menjadi 3 bagian (header, payload, signature)
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token",
		})
	}

	// Ambil header dan payload
	headerEncoded := parts[0]
	payloadEncoded := parts[1]
	signatureReceived := parts[2]

	// Verifikasi signature menggunakan HMAC SHA256
	message := headerEncoded + "." + payloadEncoded
	h := hmac.New(sha256.New, SecretKey)
	h.Write([]byte(message))
	signatureCalculated := base64.URLEncoding.EncodeToString(h.Sum(nil))

	// Bandingkan signature yang dihitung dengan signature yang diterima
	if signatureReceived != signatureCalculated {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token signature",
		})
	}

	// Decode payload untuk mengambil data pengguna
	payloadJSON, err := base64.URLEncoding.DecodeString(payloadEncoded)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token payload",
		})
	}

	var payload map[string]interface{}
	err = json.Unmarshal(payloadJSON, &payload)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token payload format",
		})
	}

	// Ambil role_id dan jenis_user_id dari payload
	roleIDHex, ok := payload["role_id"].(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Missing role_id in token",
		})
	}

	jenisUserIDHex, ok := payload["jenis_user_id"].(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Missing jenis_user_id in token",
		})
	}

	// Konversi ke primitive.ObjectID
	roleID, err := primitive.ObjectIDFromHex(roleIDHex)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid role_id in token",
		})
	}

	jenisUserID, err := primitive.ObjectIDFromHex(jenisUserIDHex)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid jenis_user_id in token",
		})
	}

	// Simpan role_id dan jenis_user_id ke context
	c.Locals("role_id", roleID)
	c.Locals("jenis_user_id", jenisUserID)
	c.Locals("username", payload["username"])

	return c.Next()
}
