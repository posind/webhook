package maxweight

import "go.mongodb.org/mongo-driver/bson/primitive"

type Item struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	DestinasiNegara  string             `bson:"DestinasiNegara" json:"DestinasiNegara"`
	KodeNegara       string             `bson:"KodeNegara" json:"KodeNegara"`
	BeratPerKoli	 string 			`bson:"BeratPerKoli" json:"BeratPerKoli"`
}
