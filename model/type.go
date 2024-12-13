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
	Data   interface{} `json:"data"`
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
	Email       string `json:"email" bson:"email"`
	Username    string `json:"username" bson:"username"`
	PhoneNumber string `bson:"phonenumber,omitempty" json:"phonenumber,omitempty"`
	Team        string `json:"team" bson:"team"`
	Scope       string `json:"scope" bson:"scope"`
	Token       string `json:"token,omitempty" bson:"token,omitempty"`
	Private     string `json:"private,omitempty" bson:"private,omitempty"`
	Public      string `json:"public,omitempty" bson:"public,omitempty"`
}

// type LoginRequest struct {
// 	PhoneNumber string `json:"phonenumber"`
// }

type Credential struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
}

// type QRStatus struct {
// 	PhoneNumber string `json:"phonenumber"`
// 	Status      bool   `json:"status"`
// 	Code        string `json:"code"`
// 	Message     string `json:"message"`
// }

// type VerifyRequest struct {
// 	PhoneNumber string `json:"phonenumber"`
// 	Password    string `json:"password"`
// }

type DestinationCode struct {
	DestinationID string `json:"destination_id" bson:"destination_id"`
	Destination   string `json:"destination" bson:"destination"`
}

type ProhibitedItems struct {
	IDItem          primitive.ObjectID `json:"id_item" bson:"_id"`
	Destination     string             `json:"destination" bson:"destination"`
	ProhibitedItems string             `json:"prohibited_items" bson:"prohibited_items"`
	MaxWeight       string             `json:"max_weight" bson:"max_weight"`
}

type Itemlarangan struct {
    IDItem          primitive.ObjectID `bson:"_id" json:"id_item"`
    Destinasi       string             `bson:"Destinasi" json:"destinasi"`
    BarangTerlarang string             `bson:"Barang Terlarang" json:"barang_terlarang"`
}

type LogInfo struct {
	PhoneNumber string `json:"phonenumber,omitempty" bson:"phonenumber,omitempty"`
	Alias       string `json:"alias,omitempty" bson:"alias,omitempty"`
	RepoOrg     string `json:"repoorg,omitempty" bson:"repoorg,omitempty"`
	RepoName    string `json:"reponame,omitempty" bson:"reponame,omitempty"`
	Commit      string `json:"commit,omitempty" bson:"commit,omitempty"`
	Remaining   int    `json:"remaining,omitempty" bson:"remaining,omitempty"`
	FileName    string `json:"filename,omitempty" bson:"filename,omitempty"`
	Base64Str   string `json:"base64str,omitempty" bson:"base64str,omitempty"`
	FileHash    string `json:"filehash,omitempty" bson:"filehash,omitempty"`
	Error       string `json:"error,omitempty" bson:"error,omitempty"`
}

type Config struct {
	PhoneNumber            string `json:"phonenumber,omitempty" bson:"phonenumber,omitempty"`
	LeaflyURL              string `json:"leaflyurl,omitempty" bson:"leaflyurl,omitempty"`
	LeaflyURLLMSDesaGambar string `json:"leaflyurllmsdesagambar,omitempty" bson:"leaflyurllmsdesagambar,omitempty"`
	LeaflyURLLMSDesaFile   string `json:"leaflyurllmsdesafile,omitempty" bson:"leaflyurllmsdesafile,omitempty"`
	LeaflySecret           string `json:"leaflysecret,omitempty" bson:"leaflysecret,omitempty"`
	DomyikadoPresensiURL   string `json:"domyikadopresensiurl,omitempty" bson:"domyikadopresensiurl,omitempty"`
	DomyikadoSecret        string `json:"domyikadosecret,omitempty" bson:"domyikadosecret,omitempty"`
	ApproveBimbinganURL    string `json:"approvebimbinganurl,omitempty" bson:"approvebimbinganurl,omitempty"`
}
