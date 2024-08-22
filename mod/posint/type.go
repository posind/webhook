package posint

import "go.mongodb.org/mongo-driver/bson/primitive"

type ItemProhibited struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Destination     string             `bson:"Destination" json:"Destination"`
	ProhibitedItems string             `bson:"Prohibited Items" json:"Prohibited Items"`
}

// type ItemLarangan struct {
// 	ID             primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
// 	Destinasi      string             `bson:"Destinasi" json:"Destinasi"`
// 	BarangTerlarang string             `bson:"Barang Terlarang" json:"Barang Terlarang"`
// }

// type ItemWeight struct {
// 	ID               primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
// 	DestinasiNegara  string             `bson:"Destinasi Negara" json:"Destinasi Negara"`
// 	KodeNegara       string             `bson:"Kode Negara" json:"Kode Negara"`
// 	BeratPerKoli	 string 			`bson:"Berat Per Koli" json:"Berat Per Koli"`
// }
