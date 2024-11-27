package controller

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gocroot/config"
	"github.com/gocroot/model"
	"github.com/kimseokgis/backend-ai/helper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetProhibitedItem(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	destination := query.Get("destination")
	prohibitedItems := query.Get("prohibited_items")

	log.Printf("Received query parameters - destination: %s, prohibited_items: %s", destination, prohibitedItems)

	filter := bson.M{}
	if destination != "" {
		filter["destination"] = destination
	}
	if prohibitedItems != "" {
		filter["prohibited_items"] = prohibitedItems
	}

	log.Printf("Filter created: %+v", filter)

	var items []model.ProhibitedItems
	collection := config.Mongoconn.Collection("prohibited_items_en")
	findOptions := options.Find()
	findOptions.SetLimit(20) // Limit hasil ke 20 dokumen

	cursor, err := collection.Find(context.Background(), filter, findOptions)
	if err != nil {
		helper.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error fetching items", "details": err.Error()})
		log.Printf("Error fetching items: %v", err)
		return
	}
	defer cursor.Close(context.Background())

	if err = cursor.All(context.Background(), &items); err != nil {
		helper.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Error decoding items", "details": err.Error()})
		log.Printf("Error decoding items: %v", err)
		return
	}

	if len(items) == 0 {
		helper.WriteJSON(w, http.StatusNotFound, "No items found")
		log.Println("No items found for the given filter")
		return
	}

	helper.WriteJSON(w, http.StatusOK, items)
	log.Printf("Successfully retrieved items: %+v", items)
}



func PostProhibitedItem(w http.ResponseWriter, r *http.Request) {
	var newItem model.ProhibitedItems
	if err := json.NewDecoder(r.Body).Decode(&newItem); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request payload", "details": err.Error()})
		log.Printf("Error decoding request payload: %v", err)
		return
	}

	// Generate a new ObjectID for the item
	newItem.IDItem = primitive.NewObjectID()

	// Validate required fields
	if newItem.Destination == "" || newItem.ProhibitedItems == "" {
		helper.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Validation error", "details": "Destination and Prohibited Items cannot be empty"})
		log.Println("Validation error: Destination or Prohibited Items is empty")
		return
	}

	// Insert item into the database
	collection := config.Mongoconn.Collection("prohibited_items_en")
	_, err := collection.InsertOne(context.Background(), newItem)
	if err != nil {
		helper.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to insert item", "details": err.Error()})
		log.Printf("Error inserting item: %v", err)
		return
	}

	helper.WriteJSON(w, http.StatusOK, newItem)
	log.Printf("Successfully added new item: %+v", newItem)
}



func UpdateProhibitedItem(w http.ResponseWriter, r *http.Request) {
	var item model.ProhibitedItems
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request payload", "details": err.Error()})
		log.Printf("Error decoding request payload: %v", err)
		return
	}

	// Validasi ObjectID
	if item.IDItem.IsZero() {
		helper.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid ID format", "details": "ID cannot be empty or invalid"})
		log.Printf("Validation error: Invalid or empty ID: %v", item.IDItem)
		return
	}

	// Define filter dan update
	filter := bson.M{"_id": item.IDItem}
	update := bson.D{ // Gunakan bson.D untuk memastikan struktur update benar
		{"$set", bson.M{
			"destination":      item.Destination,
			"prohibited_items": item.ProhibitedItems,
		}},
	}

	log.Printf("Filter: %+v, Update: %+v", filter, update)

	// Update data di database
	collection := config.Mongoconn.Collection("prohibited_items_en")
	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		helper.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to update document", "details": err.Error()})
		log.Printf("Error updating document: %v", err)
		return
	}

	if result.MatchedCount == 0 {
		helper.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "No document found to update", "details": "Check if the ID is correct"})
		log.Printf("No document found for filter: %+v", filter)
		return
	}

	helper.WriteJSON(w, http.StatusOK, map[string]string{"message": "Document updated successfully"})
	log.Printf("Successfully updated item: %+v", item)
}



func DeleteProhibitedItem(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	destination := query.Get("destination")
	prohibitedItems := query.Get("prohibited_items")

	log.Printf("Received query parameters - destination: %s, prohibited_items: %s", destination, prohibitedItems)

	filter := bson.M{}
	if destination != "" {
		filter["destination"] = destination
	}
	if prohibitedItems != "" {
		filter["prohibited_items"] = prohibitedItems
	}

	if len(filter) == 0 {
		helper.WriteJSON(w, http.StatusBadRequest, "No query parameters provided")
		log.Println("Validation error: No query parameters provided")
		return
	}

	collection := config.Mongoconn.Collection("prohibited_items_en")
	deleteResult, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		helper.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to delete item", "details": err.Error()})
		log.Printf("Error deleting document: %v", err)
		return
	}

	if deleteResult.DeletedCount == 0 {
		helper.WriteJSON(w, http.StatusNotFound, "No items found to delete")
		log.Printf("No document found for filter: %+v", filter)
		return
	}

	helper.WriteJSON(w, http.StatusOK, "Item deleted successfully")
	log.Printf("Successfully deleted item for filter: %+v", filter)
}
