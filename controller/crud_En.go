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
	// Ambil query parameter dari URL
	query := r.URL.Query()
	destinasi := query.Get("destination")          // Sesuai dengan model JSON
	barangTerlarang := query.Get("prohibited_items") // Sesuai dengan model JSON

	log.Printf("Received query parameters - destination: %s, prohibited_items: %s", destinasi, barangTerlarang)

	// Buat filter untuk query MongoDB
	filter := bson.M{}
	if destinasi != "" {
		filter["Destination"] = destinasi // Sesuaikan dengan nama field di MongoDB
	}
	if barangTerlarang != "" {
		filter["Prohibited Items"] = barangTerlarang // Sesuaikan dengan nama field di MongoDB
	}

	log.Printf("Filter created: %+v", filter)

	// Koneksi ke koleksi MongoDB
	collection := config.Mongoconn.Collection("prohibited_items_en")
	findOptions := options.Find()
	findOptions.SetLimit(20)

	// Query data dari MongoDB
	cursor, err := collection.Find(context.Background(), filter, findOptions)
	if err != nil {
		log.Printf("Error fetching items: %v", err)
		helper.WriteJSON(w, http.StatusInternalServerError, "Error fetching items")
		return
	}
	defer cursor.Close(context.Background())

	// Decode hasil query ke dalam slice model ProhibitedItems
	var items []model.ProhibitedItems
	if err = cursor.All(context.Background(), &items); err != nil {
		log.Printf("Error decoding items: %v", err)
		helper.WriteJSON(w, http.StatusInternalServerError, "Error decoding items")
		return
	}

	// Jika tidak ada data yang ditemukan
	if len(items) == 0 {
		helper.WriteJSON(w, http.StatusNotFound, "No items found")
		return
	}

	// Kirim hasil dalam format JSON
	helper.WriteJSON(w, http.StatusOK, items)
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

    // Decode JSON payload
    if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
        log.Printf("Error decoding request payload: %v", err)
        helper.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
        return
    }

    // Validasi ID
    if item.IDItem.IsZero() {
        helper.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Missing 'id_item'"})
        return
    }

    id, err := primitive.ObjectIDFromHex(item.IDItem.Hex())
    if err != nil {
        log.Printf("Invalid ObjectID: %v", err)
        helper.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid 'id_item'"})
        return
    }

    // Validasi field
    if item.Destination == "" || item.ProhibitedItems == "" {
        helper.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Missing required fields"})
        return
    }

    // Filter berdasarkan _id
    filter := bson.M{"_id": id}

    // Data yang akan diupdate
    update := bson.M{
        "$set": bson.M{
            "destination":      item.Destination,
            "prohibited_items": item.ProhibitedItems,
        },
    }

    // Eksekusi update
    collection := config.Mongoconn.Collection("prohibited_items_en")
    result, err := collection.UpdateOne(context.Background(), filter, update)
    if err != nil {
        log.Printf("Error during update: %v", err)
        helper.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Update failed"})
        return
    }

    if result.MatchedCount == 0 {
        helper.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "No matching item found"})
        return
    }

    helper.WriteJSON(w, http.StatusOK, map[string]string{"message": "Item updated successfully"})
}




func DeleteProhibitedItem(w http.ResponseWriter, r *http.Request) {
    var filter bson.M

    // Decode JSON payload untuk filter
    if err := json.NewDecoder(r.Body).Decode(&filter); err != nil {
        log.Printf("Error decoding request payload: %v", err)
        helper.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request payload", "details": err.Error()})
        return
    }

    log.Printf("Filter received: %+v", filter)

    // Validasi dan konversi _id ke ObjectID
    if id, ok := filter["_id"].(string); ok {
        objectID, err := primitive.ObjectIDFromHex(id)
        if err != nil {
            log.Printf("Invalid ObjectID format: %v", err)
            helper.WriteJSON(w, http.StatusBadRequest, "Invalid ObjectID format")
            return
        }
        filter["_id"] = objectID
    } else {
        log.Println("Missing or invalid _id in request payload")
        helper.WriteJSON(w, http.StatusBadRequest, "Missing or invalid _id in request payload")
        return
    }

    log.Printf("Filter after conversion: %+v", filter)

    // Hapus dokumen dari koleksi
    collection := config.Mongoconn.Collection("prohibited_items_en")
    deleteResult, err := collection.DeleteOne(context.Background(), filter)
    if err != nil {
        log.Printf("Error deleting items: %v", err)
        helper.WriteJSON(w, http.StatusInternalServerError, "Error deleting items")
        return
    }

    if deleteResult.DeletedCount == 0 {
        log.Println("No items found to delete")
        helper.WriteJSON(w, http.StatusNotFound, "No items found to delete")
        return
    }

    log.Println("Item deleted successfully")
    helper.WriteJSON(w, http.StatusOK, "Item deleted successfully")
}



