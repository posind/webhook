package config

import (
	"os"

	"github.com/gocroot/helper/atdb"
	"go.mongodb.org/mongo-driver/mongo"
)

var MongoString string = os.Getenv("MONGOSTRING")

var mongoinfo = atdb.DBInfo{
	DBString: MongoString,
	DBName:   "webhook",
}

var Mongoconn, ErrorMongoconn = atdb.MongoConnect(mongoinfo)

var DB *mongo.Database
