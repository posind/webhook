package gocroot

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/passwordhash"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestIsBotNumber(t *testing.T) {

}

func MongoConnect(mconn atdb.DBInfo) (db *mongo.Database) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mconn.DBString))
	if err != nil {
		fmt.Printf("AIteung Mongo, MongoConnect: %v\n", err)
	}
	return client.Database(mconn.DBName)
}

func SetConnection(MongoString, dbname string) *mongo.Database {
	MongoInfo := atdb.DBInfo{
		DBString: os.Getenv(MongoString),
		DBName:   dbname,
	}
	conn := MongoConnect(MongoInfo)
	return conn
}

// Function to compare username based on ID
func CompareUsername(MongoConn *mongo.Database, Colname, id string) bool {
	// Convert the id string to an ObjectId
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		fmt.Printf("Invalid ObjectId: %v\n", err)
		return false
	}

	filter := bson.M{"_id": objID}

	// Declare a variable to hold the result
	var user model.User

	// Fetch the document and error using GetOneDoc
	user, err = atdb.GetOneDoc[model.User](MongoConn, Colname, filter)
	if err != nil {
		fmt.Printf("Error fetching document: %v\n", err)
		return false
	}

	// Return true if the username exists in the document
	return user.Username != ""
}

// TestCompareUsername tests the CompareUsername function
func TestCompareUsername(t *testing.T) {
	// Establish the connection to MongoDB
	conn := SetConnection("MONGOSTRING", "webhook")

	// Decode the token to extract the username (assuming DecodeGetId is similar to DecodeGetUser)
	deco, err := passwordhash.DecodeGetUser("a6ffc0dc40019727d84eb9952dbd5d4cb4581d81eebca5cb3f5f42e8f1032b05",
		"v4.public.eyJleHAiOiIyMDI0LTA4LTE2VDE2OjQ0OjA2WiIsImlhdCI6IjIwMjQtMDgtMTZUMTQ6NDQ6MDZaIiwiaWQiOiI2NmJmNWU2YjlmMjBhYjZmYTYzNWY5MDEiLCJuYmYiOiIyMDI0LTA4LTE2VDE0OjQ0OjA2WiJ92XL2c7MCTRNo_3-TGIqr_RDAwgiqOXn75wZtF8BuK6Kbf8Rco9woINbvtiP3u7FF50KHEaUtNiC4HmPJ-M-MBQ")
	if err != nil {
		t.Fatalf("Failed to decode token: %v", err)
	}

	// Compare the decoded username with the ones in the database
	compare := CompareUsername(conn, "user_email", deco)

	// Print the result
	fmt.Println(compare)

	// Assert the result
	if !compare {
		t.Errorf("Expected username '%s' to be found in the database, but it was not.", deco)
	}
}
