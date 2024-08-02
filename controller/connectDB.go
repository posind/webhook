package controller

import (
	"context"
	"log"

	"github.com/gocroot/config"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectDB(user_email string) *mongo.Collection {
	clientOptions := options.Client().ApplyURI(config.MongoString)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	if err := client.Ping(context.Background(), nil); err != nil {
		log.Fatal(err)
	}
	return client.Database(user_email).Collection("user_email")
}

// SaveUserToDB saves a user to the MongoDB collection
func SaveUserToDB(user *model.User) error {
	collection := ConnectDB("user_email")
	_, err := collection.InsertOne(context.Background(), user)
	return err
}

func GetUserByEmail(email string) (model.User, error) {
	var user model.User
	collection := ConnectDB("user_email")
	err := collection.FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	return user, err
}

func GetCollection(user_email string) *mongo.Collection {
	if config.DB == nil {
		log.Fatal("Database is not initialized")
	}
	return config.DB.Collection(user_email)
}