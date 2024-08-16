package config

import (
	"log"
	"os"

	"github.com/gocroot/helper/at"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var IPPort, Net = at.GetAddress()

var PrivateKey string = os.Getenv("privateKey")

var PublicKey string = os.Getenv("publicKey")

var PhoneNumber string = os.Getenv("PHONENUMBER")

func SetEnv() {
	if ErrorMongoconn != nil {
		log.Println(ErrorMongoconn.Error())
	}
	profile, err := atdb.GetOneDoc[model.Profile](Mongoconn, "profile", primitive.M{})
	if err != nil {
		log.Println(err)
	}

	WAAPIToken = profile.Token
}
