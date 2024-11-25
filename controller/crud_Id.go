package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/at"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/watoken"
	"github.com/gocroot/model"
)

func EnsureIDitemExists(w http.ResponseWriter, r *http.Request) {
	// Temukan semua dokumen yang belum memiliki id_item atau yang memiliki id_item duplikat
	cursor, err := atdb.FindDocs(config.Mongoconn, "prohibited_items_id", bson.M{
		"$or": []bson.M{
			{"id_item": bson.M{"$exists": false}},
			{"id_item": ""},
		},
	})
	if err != nil {
		at.WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer cursor.Close(context.Background())

	batchSize := 20
	counter := 0
	var bulkWrites []mongo.WriteModel

	for cursor.Next(context.Background()) {
		var newItem model.Itemlarangan
		err := cursor.Decode(&newItem)
		if err != nil {
			at.WriteJSON(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Cari kode negara berdasarkan destinasi
		var destinationCode model.DestinationCode
		destinationCode, err = atdb.GetOneDoc[model.DestinationCode](config.Mongoconn, "destination_code", bson.M{"destination": newItem.Destinasi})
		if err != nil || destinationCode.DestinationID == "" {
			at.WriteJSON(w, http.StatusBadRequest, "Error: Could not find the country code for the given destination.")
			return
		}

		// Periksa apakah id_item sudah ada di destination yang sama dan buat ID yang unik
		isUnique := false
		itemCount := 1
		for !isUnique {
			potentialID := fmt.Sprintf("%s-%03d", destinationCode.DestinationID, itemCount)
			existingItem, err := atdb.GetOneDoc[model.Itemlarangan](config.Mongoconn, "prohibited_items_id", bson.M{
				"destination": newItem.Destinasi,
				"id_item":     potentialID,
			})
			if err != nil && err != mongo.ErrNoDocuments {
				at.WriteJSON(w, http.StatusInternalServerError, err.Error())
				return
			}
			if existingItem.IDItemIND == "" {
				newItem.IDItemIND = potentialID
				isUnique = true
			} else {
				itemCount++
			}
		}

		// Siapkan update model untuk bulk write, menggunakan filter berdasarkan _id untuk update yang tepat
		updateQuery := bson.M{
			"$set": bson.M{
				"id_item": newItem.IDItemIND,
			},
		}
		filter := bson.M{"_id": newItem.ID}
		update := mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(updateQuery)
		bulkWrites = append(bulkWrites, update)
		counter++

		// Eksekusi batch jika batchSize tercapai
		if counter >= batchSize {
			_, err := config.Mongoconn.Collection("prohibited_items_id").BulkWrite(context.Background(), bulkWrites)
			if err != nil {
				at.WriteJSON(w, http.StatusInternalServerError, err.Error())
				return
			}
			// Reset batch setelah eksekusi
			bulkWrites = nil
			counter = 0
		}
	}

	// Eksekusi batch terakhir
	if len(bulkWrites) > 0 {
		_, err := config.Mongoconn.Collection("prohibited_items_id").BulkWrite(context.Background(), bulkWrites)
		if err != nil {
			at.WriteJSON(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	at.WriteJSON(w, http.StatusOK, "Prohibited items updated successfully with new IDs where applicable.")
}

func GetItemLarangan(w http.ResponseWriter, r *http.Request) {
	var respn model.Response

	// Extract token from Login header
	tokenLogin := r.Header.Get("login")
	if tokenLogin == "" {
		respn.Status = "Error: Header Login Hilang"
		respn.Info = "Header login tidak ditemukan."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	// Decode token menggunakan DecodeGetId untuk mendapatkan private key
	privateKey, err := watoken.DecodeGetId(config.PublicKey, tokenLogin)
	if err != nil {
		respn.Status = "Error: Token Tidak Valid"
		respn.Info = "Token yang diberikan tidak valid: " + err.Error()
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	// Cari data pengguna menggunakan private key dari token
	userData, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user", bson.M{"private": privateKey})
	if err != nil || userData.PhoneNumber == "" {
		respn.Status = "Error: Tidak Berizin"
		respn.Info = "Anda tidak memiliki izin untuk mengakses data ini."
		at.WriteJSON(w, http.StatusForbidden, respn)
		return
	}

	// Ambil parameter query
	query := r.URL.Query()
	destinasi := query.Get("destinasi")
	barangTerlarang := query.Get("barang_terlarang")

	// Buat filter berdasarkan query parameter
	filterItems := bson.M{}
	if destinasi != "" {
		filterItems["destinasi"] = destinasi
	}
	if barangTerlarang != "" {
		filterItems["barang_terlarang"] = barangTerlarang
	}

	// Set limit untuk hasil
	findOptions := options.Find().SetLimit(20)
	var items []model.Itemlarangan
	collection := config.Mongoconn.Collection("prohibited_items_id")
	cursor, err := collection.Find(context.Background(), filterItems, findOptions)
	if err != nil {
		respn.Status = "Error: Internal Server Error"
		respn.Info = "Terjadi kesalahan saat mengambil data dari database."
		at.WriteJSON(w, http.StatusInternalServerError, respn)
		return
	}
	defer cursor.Close(context.Background())

	if err = cursor.All(context.Background(), &items); err != nil {
		respn.Status = "Error: Internal Server Error"
		respn.Info = "Terjadi kesalahan saat memproses data."
		at.WriteJSON(w, http.StatusInternalServerError, respn)
		return
	}
	if len(items) == 0 {
		respn.Status = "Error: Tidak Ada Data"
		respn.Info = "Tidak ada barang terlarang yang sesuai dengan filter."
		at.WriteJSON(w, http.StatusNotFound, respn)
		return
	}

	// Berikan respons dalam bentuk JSON
	at.WriteJSON(w, http.StatusOK, items)
}

func PostItemLarangan(w http.ResponseWriter, r *http.Request) {
	var respn model.Response

	// Extract token from Login header
	tokenLogin := r.Header.Get("login")
	if tokenLogin == "" {
		respn.Status = "Error: Header Login Hilang"
		respn.Info = "Header login tidak ditemukan."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	// Decode token untuk mendapatkan user ID
	userID, err := watoken.DecodeGetId(config.PublicKey, tokenLogin)
	if err != nil {
		respn.Status = "Error: Token Tidak Valid"
		respn.Info = "Token yang diberikan tidak valid: " + err.Error()
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	// Cari data pengguna menggunakan userID
	userData, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user", bson.M{"private": userID})
	if err != nil || userData.PhoneNumber == "" {
		respn.Status = "Error: Tidak Berizin"
		respn.Info = "Anda tidak memiliki izin untuk mengakses data ini."
		at.WriteJSON(w, http.StatusForbidden, respn)
		return
	}

	// Decode body permintaan untuk item baru
	var newItem model.Itemlarangan
	if err := json.NewDecoder(r.Body).Decode(&newItem); err != nil {
		at.WriteJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// Validasi input yang diperlukan
	if newItem.Destinasi == "" || newItem.BarangTerlarang == "" || newItem.IDItemIND == "" {
		at.WriteJSON(w, http.StatusBadRequest, "Semua bidang (Destinasi, Barang Terlarang, ID Item) harus diisi.")
		return
	}

	// Masukkan item ke dalam database
	collection := config.Mongoconn.Collection("prohibited_items_id")
	_, err = collection.InsertOne(context.TODO(), bson.M{
		"id_itemind":       newItem.IDItemIND,
		"destinasi":        newItem.Destinasi,
		"barang_terlarang": newItem.BarangTerlarang,
	})
	if err != nil {
		at.WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Berikan respons dengan item yang baru ditambahkan
	at.WriteJSON(w, http.StatusOK, bson.M{
		"id_itemind":       newItem.IDItemIND,
		"destinasi":        newItem.Destinasi,
		"barang_terlarang": newItem.BarangTerlarang,
	})
}

func UpdateItemLarangan(w http.ResponseWriter, r *http.Request) {
	var respn model.Response

	// Extract token from Login header
	tokenLogin := r.Header.Get("login")
	if tokenLogin == "" {
		respn.Status = "Error: Header Login Hilang"
		respn.Info = "Header login tidak ditemukan."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	// Decode token untuk mendapatkan user ID
	userID, err := watoken.DecodeGetId(config.PublicKey, tokenLogin)
	if err != nil {
		respn.Status = "Error: Token Tidak Valid"
		respn.Info = "Token yang diberikan tidak valid."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	// Cari data pengguna menggunakan userID
	userData, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user", bson.M{"private": userID})
	if err != nil || userData.PhoneNumber == "" {
		respn.Status = "Error: Tidak Berizin"
		respn.Info = "Anda tidak memiliki izin untuk mengakses data ini."
		at.WriteJSON(w, http.StatusForbidden, respn)
		return
	}

	// Decode body permintaan untuk item yang akan diperbarui
	var item model.Itemlarangan
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		at.WriteJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// Validasi bahwa id_itemind ada
	if item.IDItemIND == "" {
		at.WriteJSON(w, http.StatusBadRequest, "ID Item tidak boleh kosong.")
		return
	}

	// Definisikan filter untuk menemukan dokumen berdasarkan "id_itemind"
	filter := bson.M{"id_itemind": item.IDItemIND}

	// Definisikan bidang-bidang yang akan diperbarui
	updateFields := bson.M{
		"barang_terlarang": item.BarangTerlarang,
		"destinasi":        item.Destinasi,
	}

	// Lakukan operasi update
	result, err := atdb.UpdateOneDoc(config.Mongoconn, "prohibited_items_id", filter, updateFields)
	if err != nil {
		log.Printf("Kesalahan update MongoDB: %v", err)
		at.WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Periksa apakah ada dokumen yang diperbarui
	if result.ModifiedCount == 0 {
		respn.Status = "Error: Tidak Ada Dokumen yang Diperbarui"
		respn.Info = "Tidak ada dokumen yang diperbarui, silakan periksa id_itemind yang diberikan."
		at.WriteJSON(w, http.StatusNotFound, respn)
		return
	}

	// Berikan respons dengan item yang diperbarui
	at.WriteJSON(w, http.StatusOK, bson.M{
		"id_itemind":       item.IDItemIND,
		"destinasi":        item.Destinasi,
		"barang_terlarang": item.BarangTerlarang,
	})
}

func DeleteItemLarangan(w http.ResponseWriter, r *http.Request) {
	var respn model.Response

	// Extract token from Login header
	tokenLogin := r.Header.Get("login")
	if tokenLogin == "" {
		respn.Status = "Error: Header Login Hilang"
		respn.Info = "Header login tidak ditemukan."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	// Decode token untuk mendapatkan user ID
	userID, err := watoken.DecodeGetId(config.PublicKey, tokenLogin)
	if err != nil {
		respn.Status = "Error: Token Tidak Valid"
		respn.Info = "Token yang diberikan tidak valid."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	// Cari data pengguna menggunakan userID
	userData, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user", bson.M{"private": userID})
	if err != nil || userData.PhoneNumber == "" {
		respn.Status = "Error: Tidak Berizin"
		respn.Info = "Anda tidak memiliki izin untuk mengakses data ini."
		at.WriteJSON(w, http.StatusForbidden, respn)
		return
	}

	// Decode body permintaan
	var requestBody struct {
		IDItemIND string `json:"id_itemind"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		at.WriteJSON(w, http.StatusBadRequest, "Format JSON tidak valid")
		return
	}

	// Validasi jika id_itemind ada
	if requestBody.IDItemIND == "" {
		at.WriteJSON(w, http.StatusBadRequest, "id_itemind diperlukan")
		return
	}

	// Hapus item dari database
	_, delErr := atdb.DeleteOneDoc(config.Mongoconn, "prohibited_items_id", bson.M{"id_itemind": requestBody.IDItemIND})
	if delErr != nil {
		log.Printf("Kesalahan saat menghapus item: %v", delErr)
		at.WriteJSON(w, http.StatusInternalServerError, "Kesalahan saat menghapus item")
		return
	}

	at.WriteJSON(w, http.StatusOK, "Item berhasil dihapus")
}





