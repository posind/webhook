package maxweight

import "go.mongodb.org/mongo-driver/bson/primitive"

type Item struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	DestinasiNegara  string             `bson:"Destinasi Negara" json:"Destinasi Negara"`
	KodeNegara       string             `bson:"Kode Negara" json:"Kode Negara"`
	BeratPerKoli	 string 			`bson:"Berat Per Koli" json:"Berat Per Koli"`
}
