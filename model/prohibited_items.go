package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProhibitedItem_en struct {
    ID              primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    Destination     string             `bson:"Destination" json:"destination"`
    ProhibitedItems string             `bson:"Prohibited Items" json:"prohibited_items"`
}
type ProhibitedItem_id struct {
	ID             primitive.ObjectID  `bson:"_id,omitempty" json:"_id,omitempty"`
	Destinasi      string              `bson:"Destinasi" json:"Destinasi"`
	BrangTerlarang string              `bson:"Barang Terlarang" json:"Barang Terlarang"`
}
