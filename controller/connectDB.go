package controller

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"log"
	"os"

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

// Securely get public key from environment variable
func getRSAPublicKey() (*rsa.PublicKey, error) {
    publicKeyHex := os.Getenv("PUBLIC_KEY")
    if publicKeyHex == "" {
        return nil, errors.New("public key not found in environment variables")
    }

    publicKeyBytes, err := hex.DecodeString(publicKeyHex)
    if err != nil {
        return nil, err
    }

    pemBlock := &pem.Block{
        Bytes: publicKeyBytes,
    }

    pub, err := x509.ParsePKIXPublicKey(pemBlock.Bytes)
    if err != nil {
        return nil, err
    }

    rsaPublicKey, ok := pub.(*rsa.PublicKey)
    if !ok {
        return nil, errors.New("could not parse RSA public key")
    }

    return rsaPublicKey, nil
}