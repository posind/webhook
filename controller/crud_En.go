package controller

import (
	"context"
	"encoding/json"
	"fmt"
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

func DeleteProhibitedItem(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	destination := query.Get("destination")
	prohibitedItems := query.Get("prohibited_items")

	log.Printf("Received query parameters - destination: %s, prohibited_items: %s", destination, prohibitedItems)

	// Buat filter berdasarkan query parameters
	filter := bson.M{}
	if destination != "" {
		filter["destination"] = destination
	}
	if prohibitedItems != "" {
		filter["prohibited_items"] = prohibitedItems
	}

	// Validasi filter: Pastikan ada setidaknya satu parameter
	if len(filter) == 0 {
		helper.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "Validation error",
			"details": "At least one query parameter ('destination' or 'prohibited_items') is required",
		})
		log.Println("Validation error: No query parameters provided")
		return
	}

	log.Printf("Filter created for deletion: %+v", filter)

	// Delete dokumen menggunakan filter
	collection := config.Mongoconn.Collection("prohibited_items_en")
	deleteResult, err := collection.DeleteMany(context.Background(), filter) // Menggunakan DeleteMany untuk mendukung penghapusan banyak dokumen
	if err != nil {
		helper.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error":   "Failed to delete items",
			"details": err.Error(),
		})
		log.Printf("Error deleting documents: %v", err)
		return
	}

	// Cek apakah ada dokumen yang dihapus
	if deleteResult.DeletedCount == 0 {
		helper.WriteJSON(w, http.StatusNotFound, map[string]string{
			"message": "No items found to delete",
		})
		log.Printf("No documents found for filter: %+v", filter)
		return
	}

	// Respons jika berhasil
	helper.WriteJSON(w, http.StatusOK, map[string]string{
		"message":       "Items deleted successfully",
		"deleted_count": fmt.Sprintf("%d", deleteResult.DeletedCount),
	})
	log.Printf("Successfully deleted %d items for filter: %+v", deleteResult.DeletedCount, filter)
}


