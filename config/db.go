package config

import (
	"os"

	"github.com/gocroot/helper"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/mongo"
)

var MongoString string = os.Getenv("MONGOSTRING")

var mongoinfo = model.DBInfo{
	DBString: MongoString,
	DBName:   "webhook",
}

var Mongoconn, ErrorMongoconn = helper.MongoConnect(mongoinfo)

var DB *mongo.Database

