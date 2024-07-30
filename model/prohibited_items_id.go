package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProhibitedItem_id struct {
	ID              primitive.ObjectID  `bson:"_id,omitempty" json:"_id,omitempty"`
	Destinasi       string              `bson:"Destinasi" json:"destinasi"`
	BarangTerlarang string              `bson:"Barang Terlarang" json:"barang_terlarang"`
}
