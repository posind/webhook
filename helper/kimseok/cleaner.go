package kimseok

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func StammingQuestioninDB(mongostring, db string) {
	// Setup MongoDB connection
	clientOptions := options.Client().ApplyURI(mongostring)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.TODO())

	collection := client.Database(db).Collection("qna")

	// Find all documents in the collection
	cursor, err := collection.Find(context.TODO(), bson.D{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.TODO())

	// Iterate over documents
	for cursor.Next(context.TODO()) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			log.Fatal(err)
		}

		// Get the origin question field
		question, ok := doc["origin"].(string)
		if !ok {
			continue
		}

		// Check if 'origin' field already exists
		/* if _, exists := doc["origin"]; exists {
			// 'origin' field already exists, skip this document
			continue
		} */

		// Perform stemming
		stemmedQuestion := Stemmer(question)

		// Update document with stemmed question and origin field
		/* 		update := bson.D{
			{"$set", bson.D{
				{"question", stemmedQuestion},
				{"origin", question},
			}},
		} */

		// Update the document with the stemmed question
		filter := bson.M{"_id": doc["_id"]}
		update := bson.M{"$set": bson.M{"question": stemmedQuestion}}

		_, err := collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Updated question to: %s\n", stemmedQuestion)
	}

	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Documents updated successfully")
}
