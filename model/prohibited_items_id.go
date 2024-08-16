package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Itemlarangan struct {
    ID              primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
    Destinasi       string             `bson:"destinasi" json:"destinasi"`
    BarangTerlarang string             `bson:"Barang" json:"barang_terlarang"`
}

