package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type User_code struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Email      string             `bson:"email"`
	QRCode     string             `bson:"qr_code"`
	IsLoggedIn bool               `bson:"is_logged_in"`
}