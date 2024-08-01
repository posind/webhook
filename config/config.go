package config

import (
	"os"

	"github.com/gocroot/helper"
)

var IPPort, Net = helper.GetAddress()

var WAPhoneNumber string = "6281510040020"

var PrivateKey string = os.Getenv("privateKey")

var PublicKey string = os.Getenv("publicKey")