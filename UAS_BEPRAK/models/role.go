package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Role adalah struktur untuk data peran pengguna
type Role struct {
    ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    Name        string             `bson:"name" json:"name"`               // Nama role
    CreatedAt   primitive.DateTime `bson:"created_at" json:"created_at"`   // Tanggal pembuatan
    UpdatedAt   primitive.DateTime `bson:"updated_at" json:"updated_at"`   // Tanggal pembaruan
    CreatedBy   string             `bson:"created_by" json:"created_by"`   // Pembuat
    UpdatedBy   string             `bson:"updated_by" json:"updated_by"`   // Yang memperbarui
}
