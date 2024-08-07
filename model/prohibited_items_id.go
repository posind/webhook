package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Prohibited_items_id struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Destinasi       string             `bson:"destinasi" json:"destinasi"`
	BarangTerlarang string             `bson:"barang_terlarang" json:"barang_terlarang"`
}
