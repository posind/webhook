package kimseok

import (
	"context"
	"log"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Fungsi ini akan dijalankan oleh `go test` dan memeriksa fungsi atau metode yang ingin Anda uji.
func TestExampleFunction(t *testing.T) {
	// Setup MongoDB connection
	// Setup MongoDB connection
	mongostring := ""
	dbname := "webhook"
	clientOptions := options.Client().ApplyURI(mongostring)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.TODO())

	db := client.Database(dbname)
	q := "hari ini hari apa"
	dest, _ := QueriesDataRegexpALL(db, q)
	println(dest.Origin)
	println(dest.Question)
	println(dest.Answer)

}