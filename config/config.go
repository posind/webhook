package config

import (
	"log"
	"os"

	"github.com/gocroot/helper"
	"github.com/gocroot/helper/atdb"
	"github.com/whatsauth/itmodel"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var IPPort, Net = helper.GetAddress()

// var PrivateKey string = os.Getenv("privateKey")

// var PublicKey string = os.Getenv("publicKey")

var PhoneNumber string = os.Getenv("PHONENUMBER")

func SetEnv() {
	if ErrorMongoconn != nil {
		log.Println(ErrorMongoconn.Error())
	}
	Profile, err := atdb.GetOneDoc[itmodel.Profile](Mongoconn, "profile", primitive.M{"phonenumber": PhoneNumber})
	if err != nil {
		log.Println(err)
	}
	PublicKeyWhatsAuth = Profile.PublicKey
	WAAPIToken = Profile.Token
}
