package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Response struct {
	Response string `json:"response"`
	Info     string `json:"info,omitempty"`
	Status   string `json:"status,omitempty"`
	Location string `json:"location,omitempty"`
}

type Profile struct {
	Token       string `bson:"token"`
	Phonenumber string `bson:"phonenumber"`
	Secret      string `bson:"secret"`
	URL         string `bson:"url"`
	QRKeyword   string `bson:"qrkeyword"`
	PublicKey   string `bson:"publickey"`
}

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

type User struct {
	Name        string `json:"name" bson:"name"`
	PhoneNumber string `bson:"phonenumber,omitempty" json:"phonenumber,omitempty"`
	Team        string `json:"team" bson:"team"`
	Scope       string `json:"scope" bson:"scope"`
	Token       string `json:"token,omitempty" bson:"token,omitempty"`
	Private     string `json:"private,omitempty" bson:"private,omitempty"`
	Public      string `json:"public,omitempty" bson:"public,omitempty"`
}

type LoginRequest struct {
	PhoneNumber string `json:"phonenumber"`
}

type Credential struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
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

type DestinationCode struct {
	DestinationID string `json:"destination_id" bson:"destination_id"`
	Destination   string `json:"destination" bson:"destination"`
}

type ProhibitedItems struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	IDItem          string             `json:"id_item" bson:"id_item"`
	Destination     string             `json:"destination" bson:"Destination"`
	ProhibitedItems string             `json:"prohibited_items" bson:"Prohibited Items"`
}

type Itemlarangan struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	IDItemIND       string             `json:"id_itemind" bson:"id_itemind"`
	Destinasi       string             `bson:"destinasi" json:"destinasi"`
	BarangTerlarang string             `bson:"Barang Terlarang" json:"barang_terlarang"`
}
