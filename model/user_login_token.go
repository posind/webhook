package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Profile_user struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name,omitempty" json:"name,omitempty"`
	Email       string             `bson:"email"`
	PhoneNumber string             `bson:"phonenumber,omitempty" json:"phonenumber,omitempty"`
	QRCode      string             `bson:"qr_code"`
	Token       string             `bson:"token"`
	IsLoggedIn  bool               `bson:"is_logged_in"`
	CreatedAt   time.Time          `bson:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"`
	Private     string             `json:"private,omitempty" bson:"private,omitempty"`
	Public      string             `json:"public,omitempty" bson:"public,omitempty"`
}
type QRStatus struct {
	PhoneNumber string `json:"phonenumber"`
	Status      bool   `json:"status"`
	Code        string `json:"code"`
	Message     string `json:"message"`
}
type VerifyRequest struct {
	PhoneNumber string `json:"phonenumber"`
	Password    string `json:"password"`
}