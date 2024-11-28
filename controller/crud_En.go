package controller

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/at"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/watoken"
	"github.com/gocroot/model"
	"github.com/kimseokgis/backend-ai/helper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetProhibitedItem(w http.ResponseWriter, r *http.Request) {
	var respn model.Response

	// Ambil token login dari header menggunakan at.GetLoginFromHeader
	tokenLogin := at.GetLoginFromHeader(r)
	if tokenLogin == "" {
		respn.Status = "Error: Header Login Hilang"
		respn.Info = "Header login tidak ditemukan."
		respn.Data = nil
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		log.Println("Header login tidak ditemukan")
		return
	}

	// Log token untuk debugging
	log.Printf("Token yang diterima: %s", tokenLogin)

	// Decode token menggunakan public key dari config
	privateKey, err := watoken.DecodeGetId(config.PublicKey, tokenLogin)
	if err != nil {
		// Log kunci publik untuk memastikan kunci yang digunakan benar
		log.Printf("Kunci Publik yang digunakan: %s", config.PublicKey)
		log.Printf("Error saat validasi token: %v", err)

		respn.Status = "Error: Token Tidak Valid"
		respn.Info = "Token yang diberikan tidak valid: " + err.Error()
		respn.Data = nil
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	// Log private key yang berhasil didapatkan
	log.Printf("Private Key dari token: %s", privateKey)

	// Cari data pengguna menggunakan private key dari token
	userData, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user", bson.M{"private": privateKey})
	if err != nil || userData.PhoneNumber == "" {
		respn.Status = "Error: Tidak Berizin"
		respn.Info = "Anda tidak memiliki izin untuk mengakses data ini."
		respn.Data = nil
		at.WriteJSON(w, http.StatusForbidden, respn)
		log.Printf("Pengguna tidak memiliki izin. Private key: %s\n", privateKey)
		return
	}

	// Ambil query parameters
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

	log.Printf("Filter created: %+v", filter)

	// Inisialisasi koleksi MongoDB
	collection := config.Mongoconn.Collection("prohibited_items_en")

	// Konfigurasi opsi pencarian
	findOptions := options.Find()
	findOptions.SetLimit(20) // Batasi hasil pencarian ke 20 dokumen

	// Eksekusi pencarian
	cursor, err := collection.Find(context.Background(), filter, findOptions)
	if err != nil {
		respn.Status = "Error: Database Query"
		respn.Info = "Terjadi kesalahan saat mengambil data: " + err.Error()
		respn.Data = nil
		at.WriteJSON(w, http.StatusInternalServerError, respn)
		log.Printf("Error fetching items: %v\n", err)
		return
	}
	defer cursor.Close(context.Background())

	// Decode hasil pencarian
	var items []model.ProhibitedItems
	if err = cursor.All(context.Background(), &items); err != nil {
		respn.Status = "Error: Database Decoding"
		respn.Info = "Terjadi kesalahan saat mendecode data: " + err.Error()
		respn.Data = nil
		at.WriteJSON(w, http.StatusInternalServerError, respn)
		log.Printf("Error decoding items: %v\n", err)
		return
	}

	// Jika tidak ada data ditemukan
	if len(items) == 0 {
		respn.Status = "Not Found"
		respn.Info = "Tidak ada data yang cocok dengan filter."
		respn.Data = nil
		at.WriteJSON(w, http.StatusNotFound, respn)
		log.Println("No items found for the given filter")
		return
	}

	// Berikan respons sukses dengan data
	respn.Status = "Success"
	respn.Info = "Data berhasil diambil."
	respn.Data = items
	at.WriteJSON(w, http.StatusOK, respn)
	log.Printf("Successfully retrieved items: %+v\n", items)
}

func PostProhibitedItem(w http.ResponseWriter, r *http.Request) {
	var respn model.Response

	// Ambil token login dari header menggunakan at.GetLoginFromHeader
	tokenLogin := at.GetLoginFromHeader(r)
	if tokenLogin == "" {
		respn.Status = "Error: Header Login Hilang"
		respn.Info = "Header login tidak ditemukan."
		respn.Data = nil
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		log.Println("Header login tidak ditemukan")
		return
	}

	// Log token untuk debugging
	log.Printf("Token yang diterima: %s", tokenLogin)

	// Decode token menggunakan public key dari config
	privateKey, err := watoken.DecodeGetId(config.PublicKey, tokenLogin)
	if err != nil {
		// Log kunci publik untuk memastikan kunci yang digunakan benar
		log.Printf("Kunci Publik yang digunakan: %s", config.PublicKey)
		log.Printf("Error saat validasi token: %v", err)

		respn.Status = "Error: Token Tidak Valid"
		respn.Info = "Token yang diberikan tidak valid: " + err.Error()
		respn.Data = nil
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	// Log private key yang berhasil didapatkan
	log.Printf("Private Key dari token: %s", privateKey)

	// Cari data pengguna menggunakan private key dari token
	userData, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user", bson.M{"private": privateKey})
	if err != nil || userData.PhoneNumber == "" {
		respn.Status = "Error: Tidak Berizin"
		respn.Info = "Anda tidak memiliki izin untuk mengakses data ini."
		respn.Data = nil
		at.WriteJSON(w, http.StatusForbidden, respn)
		log.Printf("Pengguna tidak memiliki izin. Private key: %s\n", privateKey)
		return
	}

	var newItem model.ProhibitedItems
	if err := json.NewDecoder(r.Body).Decode(&newItem); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request payload", "details": err.Error()})
		log.Printf("Error decoding request payload: %v", err)
		return
	}

	newItem.IDItem = primitive.NewObjectID()

	if newItem.Destination == "" || newItem.ProhibitedItems == "" {
		helper.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Validation error", "details": "Destination and Prohibited Items cannot be empty"})
		log.Println("Validation error: Destination or Prohibited Items is empty")
		return
	}

	collection := config.Mongoconn.Collection("prohibited_items_en")
	_, err = collection.InsertOne(context.Background(), newItem)
	if err != nil {
		helper.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to insert item", "details": err.Error()})
		log.Printf("Error inserting item: %v", err)
		return
	}

	helper.WriteJSON(w, http.StatusOK, newItem)
	log.Printf("Successfully added new item: %+v", newItem)
}

func UpdateProhibitedItem(w http.ResponseWriter, r *http.Request) {
	var respn model.Response

	// Ambil token login dari header menggunakan at.GetLoginFromHeader
	tokenLogin := at.GetLoginFromHeader(r)
	if tokenLogin == "" {
		respn.Status = "Error: Header Login Hilang"
		respn.Info = "Header login tidak ditemukan."
		respn.Data = nil
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		log.Println("Header login tidak ditemukan")
		return
	}

	// Log token untuk debugging
	log.Printf("Token yang diterima: %s", tokenLogin)

	// Decode token menggunakan public key dari config
	privateKey, err := watoken.DecodeGetId(config.PublicKey, tokenLogin)
	if err != nil {
		// Log kunci publik untuk memastikan kunci yang digunakan benar
		log.Printf("Kunci Publik yang digunakan: %s", config.PublicKey)
		log.Printf("Error saat validasi token: %v", err)

		respn.Status = "Error: Token Tidak Valid"
		respn.Info = "Token yang diberikan tidak valid: " + err.Error()
		respn.Data = nil
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	// Log private key yang berhasil didapatkan
	log.Printf("Private Key dari token: %s", privateKey)

	// Cari data pengguna menggunakan private key dari token
	userData, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user", bson.M{"private": privateKey})
	if err != nil || userData.PhoneNumber == "" {
		respn.Status = "Error: Tidak Berizin"
		respn.Info = "Anda tidak memiliki izin untuk mengakses data ini."
		respn.Data = nil
		at.WriteJSON(w, http.StatusForbidden, respn)
		log.Printf("Pengguna tidak memiliki izin. Private key: %s\n", privateKey)
		return
	}

	var item model.ProhibitedItems
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request payload", "details": err.Error()})
		log.Printf("Error decoding request payload: %v", err)
		return
	}

	if item.IDItem.IsZero() {
		helper.WriteJSON(w, http.StatusBadRequest, "Invalid or missing ID")
		log.Println("Validation error: Invalid or missing ID")
		return
	}

	filter := bson.M{"_id": item.IDItem}
	update := bson.M{
		"$set": bson.M{
			"prohibited_items": item.ProhibitedItems,
			"destination":      item.Destination,
		},
	}

	log.Printf("Filter: %+v, Update: %+v", filter, update)

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
	var respn model.Response

	// Ambil token login dari header menggunakan at.GetLoginFromHeader
	tokenLogin := at.GetLoginFromHeader(r)
	if tokenLogin == "" {
		respn.Status = "Error: Header Login Hilang"
		respn.Info = "Header login tidak ditemukan."
		respn.Data = nil
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		log.Println("Header login tidak ditemukan")
		return
	}

	// Log token untuk debugging
	log.Printf("Token yang diterima: %s", tokenLogin)

	// Decode token menggunakan public key dari config
	privateKey, err := watoken.DecodeGetId(config.PublicKey, tokenLogin)
	if err != nil {
		// Log kunci publik untuk memastikan kunci yang digunakan benar
		log.Printf("Kunci Publik yang digunakan: %s", config.PublicKey)
		log.Printf("Error saat validasi token: %v", err)

		respn.Status = "Error: Token Tidak Valid"
		respn.Info = "Token yang diberikan tidak valid: " + err.Error()
		respn.Data = nil
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	// Log private key yang berhasil didapatkan
	log.Printf("Private Key dari token: %s", privateKey)

	// Cari data pengguna menggunakan private key dari token
	userData, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user", bson.M{"private": privateKey})
	if err != nil || userData.PhoneNumber == "" {
		respn.Status = "Error: Tidak Berizin"
		respn.Info = "Anda tidak memiliki izin untuk mengakses data ini."
		respn.Data = nil
		at.WriteJSON(w, http.StatusForbidden, respn)
		log.Printf("Pengguna tidak memiliki izin. Private key: %s\n", privateKey)
		return
	}

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
		log.Println("No query parameters provided. Proceeding with full collection deletion.")
	}

	collection := config.Mongoconn.Collection("prohibited_items_en")
	deleteResult, err := collection.DeleteMany(context.Background(), filter)
	if err != nil {
		helper.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to delete item", "details": err.Error()})
		log.Printf("Error deleting document: %v", err)
		return
	}

	if deleteResult.DeletedCount == 0 {
		helper.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "No items found to delete", "details": "No documents matched the filter"})
		log.Printf("No document found for filter: %+v", filter)
		return
	}

	helper.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"message":        "Item(s) deleted successfully",
		"deleted_count":  deleteResult.DeletedCount,
		"applied_filter": filter,
	})
	log.Printf("Successfully deleted %d items for filter: %+v", deleteResult.DeletedCount, filter)
}



