package alloweditems

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Allowed_Items struct {
    ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    Destination  string             `bson:"Destination" json:"destination"`
    AllowedItems string             `bson:"Allowed Items" json:"allowed_items"`
}
