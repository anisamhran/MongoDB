package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Modul adalah struktur untuk data modul
type Modul struct {
    ID             primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`          // ID unik modul
    Name           string             `bson:"name" json:"name"`                          // Nama modul
    Description    string             `bson:"description,omitempty" json:"description,omitempty"` // Deskripsi modul (opsional)
    KategoriModul  primitive.ObjectID `bson:"kategori_modul" json:"kategori_modul"`       // Referensi ke KategoriModul
    CreatedAt      primitive.DateTime `bson:"created_at" json:"created_at"` 
    CreatedBy     string             `bson:"created_by" json:"created_by"`
    UpdatedAt     primitive.DateTime `bson:"updated_at" json:"updated_at"`
    UpdatedBy     string             `bson:"updated_by" json:"updated_by"` 
    AlamatURL     string             `bson:"alamat_url" json:"alamat_url"`
    GbrIcon       string             `bson:"gbr_icon" json:"gbr_icon"`           // Waktu pembuatan
}
