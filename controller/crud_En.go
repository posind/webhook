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
		helper.WriteJSON(w, http.StatusBadRequest, "Invalid or missing ID")
		log.Println("Validation error: Invalid or missing ID")
		return
	}

	// Define filter dan update
	filter := bson.M{"_id": item.IDItem}
	update := bson.M{
		"$set": bson.M{
			"prohibited_items": item.ProhibitedItems,
			"destination":      item.Destination,
		},
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
		helper.WriteJSON(w, http.StatusNotFound, "No document found to update")
		log.Printf("No document found for filter: %+v", filter)
		return
	}

	helper.WriteJSON(w, http.StatusOK, map[string]string{"message": "Document updated successfully"})
	log.Printf("Successfully updated item: %+v", item)
}

func DeleteProhibitedItem(w http.ResponseWriter, r *http.Request) {
	// Ambil query parameters
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

	// Jika tidak ada parameter, hapus semua data (opsional, tambahkan konfirmasi jika diperlukan)
	if len(filter) == 0 {
		log.Println("No query parameters provided. Proceeding with full collection deletion.")
	}

	collection := config.Mongoconn.Collection("prohibited_items_en")

	// Hapus data berdasarkan filter
	deleteResult, err := collection.DeleteMany(context.Background(), filter) // Gunakan DeleteMany untuk penghapusan lebih fleksibel
	if err != nil {
		helper.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to delete item", "details": err.Error()})
		log.Printf("Error deleting document: %v", err)
		return
	}

	// Jika tidak ada dokumen yang dihapus
	if deleteResult.DeletedCount == 0 {
		helper.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "No items found to delete", "details": "No documents matched the filter"})
		log.Printf("No document found for filter: %+v", filter)
		return
	}

	// Berikan respons berhasil
	helper.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message":        "Item(s) deleted successfully",
		"deleted_count":  deleteResult.DeletedCount,
		"applied_filter": filter,
	})
	log.Printf("Successfully deleted %d items for filter: %+v", deleteResult.DeletedCount, filter)
}


