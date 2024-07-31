package alloweditems

import "go.mongodb.org/mongo-driver/bson/primitive"

type Item struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
    Destination  string             `bson:"Destination" json:"Destination"`
    AllowedItems string             `bson:"Allowed Items" json:"Allowed Items"`
}
