package gocroot

import (
	"fmt"
	"testing"

	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/passwordhash"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestIsBotNumber(t *testing.T) {

}

// Function to compare username based on username
func CompareUsername(MongoConn *mongo.Database, Colname, username string) bool {
	// Create a filter using the username
	filter := bson.M{"username": username}

	// Declare a variable to hold the result
	var user model.User

	// Fetch the document and error using GetOneDoc
	user, err := atdb.GetOneDoc[model.User](MongoConn, Colname, filter)
	if err != nil {
		fmt.Printf("Error fetching document: %v\n", err)
		return false
	}

	// Return true if the Name exists in the document
	return user.Name != ""
}

// TestCompareUsername tests the CompareUsername function
func TestCompareUsername(t *testing.T) {
	// Define your MongoDB connection info
	dbInfo := atdb.DBInfo{
		DBString: "your_mongodb_connection_string", // Replace with your MongoDB connection string
		DBName:   "webhook",                        // Replace with your database name
	}

	// Establish the connection to MongoDB using MongoConnect
	conn, err := atdb.MongoConnect(dbInfo)
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Decode the token to extract the username (assuming DecodeGetUser returns a username)
	deco, err := passwordhash.DecodeGetUser("a6ffc0dc40019727d84eb9952dbd5d4cb4581d81eebca5cb3f5f42e8f1032b05",
		"v4.public.eyJleHAiOiIyMDI0LTA4LTE2VDE4OjM4OjQyWiIsImlhdCI6IjIwMjQtMDgtMTZUMTY6Mzg6NDJaIiwiaWQiOiJmYWhpcmExNCIsIm5iZiI6IjIwMjQtMDgtMTZUMTY6Mzg6NDJaIn0FkKJlWFd-a_7o6evBvfjFO45SHq9hFFav6wLT_k89S6S1hbe2JUhhTpPs0SbdoxumKJnf91KAfZDNJroJVKoN")
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
