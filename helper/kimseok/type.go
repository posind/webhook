package kimseok

import "go.mongodb.org/mongo-driver/bson/primitive"

type Datasets struct {
	ID       primitive.ObjectID `json:"id" bson:"_id"`
	Question string             `json:"question" bson:"question"`
	Answer   string             `json:"answer" bson:"answer"`
	Origin   string             `json:"origin" bson:"origin"`
}

// Struktur untuk menyimpan nama negara
type Country struct {
	NamaNegara  string `json:"namanegara"`
	CountryName string `json:"countryname"`
}

// Definisikan struct untuk data MongoDB id
type DestinasiTerlarang struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Destinasi       string             `bson:"Destinasi" json:"Destinasi"`
	BarangTerlarang string             `bson:"Barang Terlarang" json:"Barang Terlarang"`
}

// Definisikan struct untuk data MongoDB en
type DestinationProhibit struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Destination     string             `bson:"Destination" json:"Destination"`
	ProhibitedItems string             `bson:"Prohibited Items" json:"Prohibited Items"`
}

// Struktur untuk berat max
type MaxWeight struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	DestinasiNegara  string             `bson:"Destinasi Negara" json:"Destinasi Negara"`
	KodeNegara       string             `bson:"Kode Negara" json:"Kode Negara"`
	BeratPerKoli	 string 			`bson:"Berat Per Koli" json:"Berat Per Koli"`
}
