package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserModul adalah struktur untuk menyimpan hubungan antara user dan modul
type UserModul struct {
	ModulID    primitive.ObjectID `json:"modul_id" bson:"modul_id"`       // Referensi ke Modul
	NamaModul	string				`json:"nm_modul" bson:"nm_modul"` 
	CreatedAt  primitive.DateTime `json:"created_at" bson:"created_at"`   // Waktu pembuatan
	CreatedBy  string             `json:"created_by" bson:"created_by"`   // User yang membuat
	UpdatedAt  primitive.DateTime `json:"updated_at" bson:"updated_at"`   // Waktu terakhir diperbarui
	UpdatedBy  string             `json:"updated_by" bson:"updated_by"`   // User yang memperbarui
}

// User adalah struktur untuk data pengguna
type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`         // ID unik pengguna
	Username     string             `bson:"username" json:"username"`                 // Username
	NmUser       string             `bson:"nm_user" json:"nm_user"`                   // Nama user
	Pass         string             `bson:"pass" json:"-"`                            // Password (hashed)
	Email        string             `bson:"email" json:"email"`                       // Email pengguna
	Photo         string             `bson:"photo" json:"photo"`
    Phone         string             `bson:"phone" json:"phone"`
    RoleID       primitive.ObjectID `bson:"role_id" json:"role_id"`                   // Referensi ke Role
	JenisKelamin string             `bson:"jenis_kelamin" json:"jenis_kelamin"`       // Jenis kelamin
	JenisUserID  primitive.ObjectID `bson:"jenis_user_id" json:"jenis_user_id"`       // Referensi ke JenisUser
	Token        string             `bson:"token,omitempty" json:"token,omitempty"`   // Token JWT (opsional)
	UserModul    []UserModul        `bson:"user_modul" json:"user_modul"`             // Modul yang diakses oleh user
	CreatedAt    primitive.DateTime `bson:"created_at" json:"created_at"`             // Waktu pembuatan
    CreatedBy     string             `bson:"created_by" json:"created_by"`
    UpdatedAt     primitive.DateTime `bson:"updated_at" json:"updated_at"`
    UpdatedBy     string             `bson:"updated_by" json:"updated_by"`
}