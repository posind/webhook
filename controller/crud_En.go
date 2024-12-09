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
    var requestBody map[string]interface{}

    // Decode JSON payload menjadi map
    if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
        log.Printf("Error decoding request payload: %v", err)
        helper.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request payload", "details": err.Error()})
        return
    }

    log.Printf("Request body received: %+v", requestBody)

    // Validasi `id_item`
    idItem, ok := requestBody["id_item"].(string)
    if !ok || idItem == "" {
        log.Println("Missing or invalid 'id_item'")
        helper.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Validation error", "details": "Missing or invalid 'id_item'"})
        return
    }

    objectID, err := primitive.ObjectIDFromHex(idItem)
    if err != nil {
        log.Printf("Invalid ObjectID format for 'id_item': %v", err)
        helper.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid ObjectID format", "details": err.Error()})
        return
    }

    // Validasi `destination`
    destination, ok := requestBody["destination"].(string)
    if !ok || destination == "" {
        log.Println("Missing or invalid 'destination'")
        helper.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Validation error", "details": "Missing or invalid 'destination'"})
        return
    }

    // Validasi `prohibited_items`
    prohibitedItems, ok := requestBody["prohibited_items"].(string)
    if !ok || prohibitedItems == "" {
        log.Println("Missing or invalid 'prohibited_items'")
        helper.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Validation error", "details": "Missing or invalid 'prohibited_items'"})
        return
    }

    // Buat filter dan update
    filter := bson.M{"_id": objectID}
    update := bson.M{
        "$set": bson.M{
            "destination":      destination,
            "prohibited_items": prohibitedItems,
        },
    }

    log.Printf("Filter: %+v, Update: %+v", filter, update)

    // Update dokumen di koleksi
    collection := config.Mongoconn.Collection("prohibited_items_en")
    updateResult, err := collection.UpdateOne(context.Background(), filter, update)
    if err != nil {
        log.Printf("Error updating item: %v", err)
        helper.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to update item", "details": err.Error()})
        return
    }

    if updateResult.ModifiedCount == 0 {
        log.Println("No items found to update")
        helper.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "No items found to update"})
        return
    }

    log.Println("Item updated successfully")
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



