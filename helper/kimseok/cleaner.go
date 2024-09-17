package kimseok

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func StammingQuestioninDB(mongostring, db string) error {
	// Setup MongoDB connection
	clientOptions := options.Client().ApplyURI(mongostring)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(context.TODO())

	collection := client.Database(db).Collection("ssd")

	// Find all documents in the collection
	cursor, err := collection.Find(context.TODO(), bson.D{})
	if err != nil {
		return fmt.Errorf("failed to retrieve documents: %v", err)
	}
	defer cursor.Close(context.TODO())

	// Iterate over documents
	var bulkOps []mongo.WriteModel
	for cursor.Next(context.TODO()) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return fmt.Errorf("failed to decode document: %v", err)
		}

		// Get the origin question field
		question, ok := doc["origin"].(string)
		if !ok {
			// Skip document if no "origin" field
			continue
		}

		// Perform stemming
		stemmedQuestion := Stemmer(question)

		// Prepare bulk update
		filter := bson.M{"_id": doc["_id"]}
		update := bson.M{"$set": bson.M{"question": stemmedQuestion, "origin": question}}

		bulkOps = append(bulkOps, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update))
	}

	if len(bulkOps) > 0 {
		// Execute bulk write operation
		bulkOpts := options.BulkWrite().SetOrdered(true)
		_, err := collection.BulkWrite(context.TODO(), bulkOps, bulkOpts)
		if err != nil {
			return fmt.Errorf("bulk update failed: %v", err)
		}
		fmt.Printf("Documents updated successfully: %d\n", len(bulkOps))
	}

	// Check for any errors after iteration
	if err := cursor.Err(); err != nil {
		return fmt.Errorf("cursor error: %v", err)
	}

	return nil
}
