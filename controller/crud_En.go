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
)

func GetProhibitedItem(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	destination := query.Get("destination")
	prohibitedItems := query.Get("prohibited_items")

	log.Printf("Received query parameters - destination: %s, prohibited_items: %s", destination, prohibitedItems)

	// Buat filter berdasarkan parameter
	filter := bson.M{}
	if destination != "" {
		filter["destination"] = destination
	}
	if prohibitedItems != "" {
		filter["prohibited_items"] = prohibitedItems
	}

	log.Printf("Filter created: %+v", filter)

	// Query ke MongoDB
	var items []model.ProhibitedItems
	collection := config.Mongoconn.Collection("prohibited_items_en")
	cursor, err := collection.Find(context.Background(), filter) // Tidak ada limit, ambil semua data yang cocok
	if err != nil {
		helper.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error":   "Error fetching items",
			"details": err.Error(),
		})
		log.Printf("Error fetching items: %v", err)
		return
	}
	defer cursor.Close(context.Background())

	// Decode hasil query ke dalam slice
	if err := cursor.All(context.Background(), &items); err != nil {
		helper.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error":   "Error decoding items",
			"details": err.Error(),
		})
		log.Printf("Error decoding items: %v", err)
		return
	}

	// Cek apakah hasil kosong
	if len(items) == 0 {
		helper.WriteJSON(w, http.StatusNotFound, map[string]string{
			"message": "No items found",
		})
		log.Println("No items found for the given filter")
		return
	}

	helper.WriteJSON(w, http.StatusOK, items)
	log.Printf("Successfully retrieved items: %d items", len(items))
}



func PostProhibitedItem(w http.ResponseWriter, r *http.Request) {
	var newItem model.ProhibitedItems

	// Decode payload JSON
	if err := json.NewDecoder(r.Body).Decode(&newItem); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request payload", "details": err.Error()})
		log.Printf("Error decoding request payload: %v", err)
		return
	}

	// Validasi data wajib
	if newItem.Destination == "" || newItem.ProhibitedItems == "" {
		helper.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Validation error", "details": "Destination and Prohibited Items cannot be empty"})
		log.Println("Validation error: Destination or Prohibited Items is empty")
		return
	}

	// Generate ObjectID
	newItem.IDItem = primitive.NewObjectID()

	// Insert ke database
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

	// Decode payload JSON
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request payload", "details": err.Error()})
		log.Printf("Error decoding request payload: %v", err)
		return
	}

	// Validasi ObjectID
	if item.IDItem.IsZero() {
		helper.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Validation error", "details": "Invalid or missing ID"})
		log.Println("Validation error: Invalid or missing ID")
		return
	}

	// Filter dan update
	filter := bson.M{"_id": item.IDItem}
	update := bson.M{
		"$set": bson.M{
			"destination":      item.Destination,
			"prohibited_items": item.ProhibitedItems,
		},
	}

	log.Printf("Filter: %+v, Update: %+v", filter, update)

	// Update database
	collection := config.Mongoconn.Collection("prohibited_items_en")
	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		helper.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to update item", "details": err.Error()})
		log.Printf("Error updating item: %v", err)
		return
	}

	if result.ModifiedCount == 0 {
		helper.WriteJSON(w, http.StatusNotFound, map[string]string{"message": "No document found to update"})
		log.Printf("No document found for filter: %+v", filter)
		return
	}

	helper.WriteJSON(w, http.StatusOK, item)
	log.Printf("Successfully updated item: %+v", item)
}

// DeleteProhibitedItemByField deletes an item based on provided fields.
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

	log.Printf("Filter created: %+v", filter)

	if len(filter) == 0 {
		log.Println("No query parameters provided, returning all items.")
	}

	collection := config.Mongoconn.Collection("prohibited_items_en")
	deleteResult, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		log.Printf("Error deleting items: %v", err)
		helper.WriteJSON(w, http.StatusInternalServerError, "Error deleting items")
		return
	}

	if deleteResult.DeletedCount == 0 {
		helper.WriteJSON(w, http.StatusNotFound, "No items found to delete")
		return
	}

	helper.WriteJSON(w, http.StatusOK, "Item deleted successfully")
}

