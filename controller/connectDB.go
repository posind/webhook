package controller

import (
	"context"
	"log"

	"github.com/gocroot/config"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func ConnectDB(user_email string) *mongo.Collection {
	return (config.Mongoconn).Collection(user_email)
}

// SaveUserToDB saves a user to the MongoDB collection
func SaveUserToDB(user *model.User) error {
	collection := ConnectDB("user_email")
	_, err := collection.InsertOne(context.Background(), user)
	if err != nil {
		log.Printf("Error inserting user to DB: %v", err)
	}
	return err
}

func GetUserByEmail(email string) (model.User, error) {
	var user model.User
	collection := ConnectDB("user_email")
	err := collection.FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		log.Printf("Error finding user by email: %v", err)
	}
	return user, err
}

