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
	findOptions.SetLimit(20)
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

// Fungsi untuk memperbarui item larangan
func UpdateItemLarangan(w http.ResponseWriter, r *http.Request) {
	var item model.Itemlarangan

	// Decode JSON
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		at.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "Payload tidak valid",
			"details": err.Error(),
		})
		log.Printf("Kesalahan saat decode payload: %v", err)
		return
	}

	// Validasi ObjectID
	if item.IDItem.IsZero() {
		at.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "Validasi gagal",
			"details": "ID tidak valid atau kosong",
		})
		log.Println("Validasi gagal: ID tidak valid atau kosong")
		return
	}

	// Filter dan update
	filter := bson.M{"_id": item.IDItem}
	update := bson.M{
		"$set": bson.M{
			"destinasi":       item.Destinasi,
			"barang_terlarang": item.BarangTerlarang,
		},
	}

	log.Printf("Filter: %+v, Update: %+v", filter, update)

	// Update database
	collection := config.Mongoconn.Collection("prohibited_items_id")
	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		at.WriteJSON(w, http.StatusInternalServerError, map[string]string{
			"error":   "Gagal memperbarui data",
			"details": err.Error(),
		})
		log.Printf("Kesalahan saat memperbarui data: %v", err)
		return
	}

	if result.ModifiedCount == 0 {
		at.WriteJSON(w, http.StatusNotFound, map[string]string{
			"message": "Tidak ada dokumen yang diperbarui",
		})
		log.Printf("Tidak ada dokumen yang sesuai dengan filter: %+v", filter)
		return
	}

	at.WriteJSON(w, http.StatusOK, item)
	log.Printf("Berhasil memperbarui item: %+v", item)
}

// DeleteProhibitedItemByField deletes an item based on provided fields.
func DeleteItemLarangan(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	destinasi := query.Get("destinasi")
	barangTerlarang := query.Get("barang_terlarang")

	log.Printf("Received query parameters - destinasi: %s, barang_terlarang: %s", destinasi, barangTerlarang)

	filter := bson.M{}
	if destinasi != "" {
		filter["destinasi"] = destinasi
	}
	if 	barangTerlarang != "" {
		filter["barang_terlarang"] = barangTerlarang
	}

	log.Printf("Filter created: %+v", filter)

	if len(filter) == 0 {
		log.Println("No query parameters provided, returning all items.")
	}

	collection := config.Mongoconn.Collection("prohibited_items_id")
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





