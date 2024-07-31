package config

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gocroot/helper"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoString string = os.Getenv("MONGOSTRING")

var mongoinfo = model.DBInfo{
	DBString: MongoString,
	DBName:   "webhook",
}

var Mongoconn, ErrorMongoconn = helper.MongoConnect(mongoinfo)

var DB *mongo.Database

func InitDB() {
	clientOptions := options.Client().ApplyURI("mongodb+srv://mfulbiposin:U5lTmGfG9C0FYvBL@cluster0.nxjo2cg.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0")
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	DB = client.Database("mfulbiposin")
	log.Println("Connected to MongoDB!")
}

func GetCollection(user_email string) *mongo.Collection {
    if DB == nil {
        log.Fatal("Database is not initialized")
    }
    return DB.Collection(user_email)
}