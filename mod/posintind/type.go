package posintind

import "go.mongodb.org/mongo-driver/bson/primitive"

type Item struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Destinasi      string             `bson:"Destinasi" json:"Destinasi"`
	BarangTerlarang string             `bson:"Barang Terlarang" json:"Barang Terlarang"`
}
