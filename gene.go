package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FAQ represents a MongoDB document
type FAQ struct {
	Question string `bson:"question"`
	Answer   string `bson:"answer"`
	Origin   string `bson:"origin"`
}

func main() {
	// Seed the random generator
	gofakeit.Seed(time.Now().UnixNano())

	// Connect to MongoDB
	clientOptions := options.Client().ApplyURI("mongodb+srv://mfulbiposin:U5lTmGfG9C0FYvBL@cluster0.nxjo2cg.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.TODO())

	// Access the 'mf' database and 'faq' collection
	db := client.Database("mf")
	faqCollection := db.Collection("faq")

	// Generate random FAQs
	faqs := generateFAQs(20)

	// Insert the generated dataset into MongoDB
	var docs []interface{}
	for _, faq := range faqs {
		docs = append(docs, faq)
	}

	result, err := faqCollection.InsertMany(context.TODO(), docs)
	if err != nil {
		log.Fatal(err)
	}

	// Print the number of inserted documents
	fmt.Printf("Inserted %d documents into MongoDB.\n", len(result.InsertedIDs))

	// Retrieve and print the inserted data
	cursor, err := faqCollection.Find(context.TODO(), bson.D{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var faq FAQ
		err := cursor.Decode(&faq)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(faq)
	}
}

// generateFAQs generates a list of random FAQs
func generateFAQs(num int) []FAQ {
	var faqs []FAQ
	for i := 0; i < num; i++ {
		country := gofakeit.Country()                            // Generate random country
		item := gofakeit.Word()                                  // Generate random item (using gofakeit)
		question := fmt.Sprintf("%s %s", country, item)          // Generate the question
		answer := fmt.Sprintf("Pengiriman %s ke %s dilarang.", item, country) // Generate the answer

		faq := FAQ{
			Question: question,
			Answer:   answer,
			Origin:   "menu",
		}
		faqs = append(faqs, faq)
	}
	return faqs
}
