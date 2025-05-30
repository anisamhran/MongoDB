package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TemplateModul adalah struktur untuk menyimpan hubungan antara jenis user dan modul
type TemplateModul struct {
    ModulID primitive.ObjectID `json:"modul_id" bson:"modul_id"` // Referensi ke modul
}

// JenisUser adalah struktur untuk menyimpan jenis user dengan template modul yang di-*embed*
type JenisUser struct {
    ID          primitive.ObjectID  `json:"id" bson:"_id"`            // ID unik untuk jenis user
    NmJenisUser string              `json:"nm_jenis_user" bson:"nm_jenis_user"` // Nama jenis user
    TemplateModul []TemplateModul   `json:"template_modul" bson:"template_modul"` // Daftar modul yang di-embed
    CreatedAt   primitive.DateTime `json:"created_at" bson:"created_at"` // Waktu pembuatan
    UpdatedAt   primitive.DateTime `json:"updated_at" bson:"updated_at"` // Waktu terakhir diperbarui
    CreatedBy   string              `json:"created_by" bson:"created_by"` // User yang membuat
    UpdatedBy   string              `json:"updated_by" bson:"updated_by"` // User yang memperbarui
}
