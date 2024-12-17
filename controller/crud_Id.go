package controller

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/at"
	"github.com/gocroot/model"
	"github.com/kimseokgis/backend-ai/helper"
)

func GetItemLarangan(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	destinasi := query.Get("destinasi")
	barangTerlarang := query.Get("barang_terlarang") // Sesuai dengan field JSON

	log.Printf("Received query parameters - destinasi: %s, barang_terlarang: %s", destinasi, barangTerlarang)

	// Buat filter berdasarkan query parameter
	filter := bson.M{}
	if destinasi != "" {
		filter["Destinasi"] = destinasi // Sesuai dengan field di database
	}
	if barangTerlarang != "" {
		filter["Barang Terlarang"] = barangTerlarang // Sesuai dengan field di database
	}

	log.Printf("Filter created: %+v", filter)

	// Koneksi ke koleksi MongoDB
	var items []model.Itemlarangan
	collection := config.Mongoconn.Collection("prohibited_items_id")
	findOptions := options.Find()
	findOptions.SetLimit(100)
	cursor, err := collection.Find(context.Background(), filter, findOptions)
	if err != nil {
		helper.WriteJSON(w, http.StatusInternalServerError, "Error fetching items")
		return
	}
	defer cursor.Close(context.Background())

	// Decode data ke model
	for cursor.Next(context.Background()) {
		var raw bson.M
		if err := cursor.Decode(&raw); err != nil {
			log.Printf("Error decoding raw item: %v", err)
			continue
		}

		// Map field MongoDB ke model
		item := model.Itemlarangan{
			IDItem: raw["_id"].(primitive.ObjectID),
			Destinasi: func() string {
				if dest, ok := raw["Destinasi"].(string); ok { //sesuaikan dengan field di database
					return dest
				}
				return ""
			}(),
			BarangTerlarang: func() string {
				if prohibited, ok := raw["Barang Terlarang"].(string); ok { //sesuaikan dengan field dalam database
					return prohibited
				}
				return ""
			}(),
		}
		items = append(items, item)
	}

	// Cek jika tidak ada item yang ditemukan
	if len(items) == 0 {
		helper.WriteJSON(w, http.StatusNotFound, "No items found")
		return
	}

	// Tampilkan hasil dalam JSON
	helper.WriteJSON(w, http.StatusOK, items)
}


// Fungsi untuk menambahkan item larangan
func PostItemLarangan(w http.ResponseWriter, r *http.Request) {
	var itemBaru model.Itemlarangan

	// Decode JSON
	if err := json.NewDecoder(r.Body).Decode(&itemBaru); err != nil {
		at.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "Payload tidak valid",
			"details": err.Error(),
		})
		log.Printf("Kesalahan saat decode payload: %v", err)
		return
	}

	// Validasi data wajib
	if itemBaru.Destinasi == "" || itemBaru.BarangTerlarang == "" {
		at.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "Validasi gagal",
			"details": "Destinasi dan Barang Terlarang tidak boleh kosong",
		})
		log.Println("Validasi gagal: Destinasi atau Barang Terlarang kosong")
		return
	}

	// Generate ObjectID
	itemBaru.IDItem = primitive.NewObjectID()

	// Insert ke database
	collection := config.Mongoconn.Collection("prohibited_items_id")
	_, err := collection.InsertOne(context.Background(), itemBaru)
	if err != nil {
		at.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error":   "Gagal menambahkan data",
			"details": err.Error(),
		})
		log.Printf("Kesalahan saat menambahkan data: %v", err)
		return
	}

	at.WriteJSON(w, http.StatusOK, itemBaru)
	log.Printf("Berhasil menambahkan item baru: %+v", itemBaru)
}

//Update
func UpdateItemLarangan(w http.ResponseWriter, r *http.Request) {
    var item model.Itemlarangan

    // Decode JSON payload
    if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
        log.Printf("Error decoding request payload: %v", err)
        helper.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request payload", "details": err.Error()})
        return
    }

    log.Printf("Received payload: %+v", item) // Log tambahan

    // Validasi ID (id_item harus valid ObjectID)
    if item.IDItem.IsZero() {
        helper.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Missing 'id_item'", "details": "Field 'id_item' is required"})
        return
    }

    id, err := primitive.ObjectIDFromHex(item.IDItem.Hex())
    if err != nil {
        log.Printf("Invalid 'id_item': %v", err)
        helper.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid 'id_item'", "details": "ObjectID format is invalid"})
        return
    }
    item.IDItem = id

    // Validasi field lainnya
    if item.Destinasi == "" || item.BarangTerlarang == "" {
        helper.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Validation error", "details": "Missing required fields"})
        return
    }

    // Filter dan Update
    filter := bson.M{"_id": item.IDItem}
    update := bson.M{
        "$set": bson.M{
            "Destinasi":        item.Destinasi,
            "Barang Terlarang": item.BarangTerlarang,
        },
    }

    // Update MongoDB
    collection := config.Mongoconn.Collection("prohibited_items_id")
    result, err := collection.UpdateOne(context.Background(), filter, update)
    if err != nil {
        log.Printf("Error updating item: %v", err)
        helper.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to update item", "details": err.Error()})
        return
    }

    if result.MatchedCount == 0 {
        helper.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "No items found", "details": "No matching document found"})
        return
    }

    helper.WriteJSON(w, http.StatusOK, map[string]string{"message": "Item updated successfully"})
}



// DeleteProhibitedItem deletes an item based on provided fields.
func DeleteItemLarangan(w http.ResponseWriter, r *http.Request) {
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
    collection := config.Mongoconn.Collection("prohibited_items_id")
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





