package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Allowed_Items struct {
    ID              primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    Destination     string             `bson:"Destination" json:"destination"`
    ProhibitedItems string             `bson:"Prohibited Items" json:"prohibited_items"`
}