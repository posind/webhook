package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/exp/rand"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/at"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/passwordhash"
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


// GetItemIND - Mengambil item terlarang
func GetitemIND(w http.ResponseWriter, r *http.Request) {
	var respn model.Response

	// Ambil token dari header Authorization
	authHeader := r.Header.Get("Authorization")
	tokenLogin := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenLogin == "" {
		respn.Status = "Error: Missing Authorization header"
		respn.Info = "Authorization header is missing or invalid."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	// Temukan user berdasarkan token di database
	userData, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user", bson.M{"token": tokenLogin})
	if err != nil || userData.PhoneNumber == "" {
		log.Printf("Error finding user by token: %v", err)
		respn.Status = "Error: Tidak Diizinkan"
		respn.Info = "Anda tidak memiliki izin untuk mengakses data ini."
		at.WriteJSON(w, http.StatusForbidden, respn)
		return
	}

	// Debugging log
	log.Printf("User found: %s", userData.PhoneNumber)

	// Dekode token menggunakan kunci publik user
	decodedPhoneNumber, err := passwordhash.DecodeGetUser(userData.Public, tokenLogin)
	if err != nil {
		log.Printf("Error decoding token: %v", err)
		respn.Status = "Error: Token Tidak Valid"
		respn.Info = "Token yang diberikan tidak valid."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	// Debugging log
	log.Printf("Token decoded, phone number: %s", decodedPhoneNumber)

	// Periksa apakah nomor telepon yang terdekode ada di database
	userByPhoneNumber, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user", bson.M{"phonenumber": decodedPhoneNumber})
	if err != nil || userByPhoneNumber.PhoneNumber == "" {
		log.Printf("User not found by phone number: %s", decodedPhoneNumber)
		respn.Status = "Error: User Tidak Ditemukan"
		respn.Info = fmt.Sprintf("Nomor telepon '%s' yang diambil dari token tidak ada di database.", decodedPhoneNumber)
		at.WriteJSON(w, http.StatusForbidden, respn)
		return
	}

	// Token valid, lanjutkan mengambil data
	query := r.URL.Query()
	destination := query.Get("destination")
	prohibitedItems := query.Get("prohibited_items_id")

	// Debugging log for query parameters
	log.Printf("Query parameters - destination: %s, prohibited_items: %s", destination, prohibitedItems)

	// Build the filter for MongoDB query
	filterItems := bson.M{}
	if destination != "" {
		filterItems["destination"] = destination
	}
	if prohibitedItems != "" {
		filterItems["Prohibited Items"] = prohibitedItems
	}

	findOptions := options.Find().SetLimit(20)

	// Query ke MongoDB
	var items []model.Itemlarangan
	collection := config.Mongoconn.Collection("prohibited_items_id")

	cursor, err := collection.Find(context.Background(), filterItems, findOptions)
	if err != nil {
		log.Printf("Error querying MongoDB: %v", err)
		respn.Status = "Error: Kesalahan Server Internal"
		respn.Info = "Kesalahan saat mengambil item dari database."
		at.WriteJSON(w, http.StatusInternalServerError, respn)
		return
	}
	defer cursor.Close(context.Background())

	// Decode results from MongoDB cursor
	if err = cursor.All(context.Background(), &items); err != nil {
		log.Printf("Error decoding MongoDB result: %v", err)
		respn.Status = "Error: Kesalahan Saat Mendekode Data"
		respn.Info = "Kesalahan mendekode data item."
		at.WriteJSON(w, http.StatusInternalServerError, respn)
		return
	}

	// If no items are found
	if len(items) == 0 {
		log.Printf("No items found for the given filter")
		respn.Status = "Error: Tidak Ada Item Ditemukan"
		respn.Info = "Tidak ada item yang cocok dengan filter yang diberikan."
		at.WriteJSON(w, http.StatusNotFound, respn)
		return
	}

	// Balas dengan item dalam format JSON
	at.WriteJSON(w, http.StatusOK, items)
}


func PostitemIND(w http.ResponseWriter, r *http.Request) {
    var respn model.Response

    // Ambil token dari header Authorization
	authHeader := r.Header.Get("Authorization")
	tokenLogin := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenLogin == "" {
		respn.Status = "Error: Missing Authorization header"
		respn.Info = "Authorization header is missing or invalid."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

    // Temukan user berdasarkan token di database
    userData, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user", bson.M{"token": tokenLogin})
    if err != nil || userData.PhoneNumber == "" {
        respn.Status = "Error: Tidak Diizinkan"
        respn.Info = "Anda tidak memiliki izin untuk mengakses data ini."
        at.WriteJSON(w, http.StatusForbidden, respn)
        return
    }

    // Decode request body untuk item baru
    var newItem model.Itemlarangan
    if err := json.NewDecoder(r.Body).Decode(&newItem); err != nil {
        at.WriteJSON(w, http.StatusBadRequest, err.Error())
        return
    }

    // Validasi input yang diperlukan
    if newItem.Destinasi == "" {
        at.WriteJSON(w, http.StatusBadRequest, "Destinasi tidak boleh kosong.")
        return
    }
    if newItem.BarangTerlarang == "" {
        at.WriteJSON(w, http.StatusBadRequest, "Barang Terlarang tidak boleh kosong.")
        return
    }

    // Buat tiga digit acak untuk ID Item
    randomDigits := fmt.Sprintf("%03d", rand.Intn(1000))

    // Cari kode destinasi di database
    var destinationCode model.DestinationCode
    destinationCode, err = atdb.GetOneDoc[model.DestinationCode](config.Mongoconn, "destination_code", bson.M{"destination": newItem.Destinasi})
    if err != nil || destinationCode.DestinationID == "" {
        respn.Status = "Error: Destinasi Tidak Valid"
        respn.Info = "Kode negara untuk destinasi yang diberikan tidak ditemukan."
        at.WriteJSON(w, http.StatusBadRequest, respn)
        return
    }

    // Buat id_item otomatis
    newItem.IDItemIND = fmt.Sprintf("%s-%s", destinationCode.DestinationID, randomDigits)

    // Buat dokumen yang hanya berisi field yang diperlukan
    document := bson.M{
        "id_itemind":       newItem.IDItemIND,
        "destinasi":        newItem.Destinasi,
        "barang_terlarang": newItem.BarangTerlarang,
    }

    // Masukkan data item ke database
    collection := config.Mongoconn.Collection("prohibited_items_id")
    _, err = collection.InsertOne(context.TODO(), document)
    if err != nil {
        at.WriteJSON(w, http.StatusInternalServerError, err.Error())
        return
    }

    // Berhasil menambahkan item
    at.WriteJSON(w, http.StatusOK, document)
}


func UpdateitemIND(w http.ResponseWriter, r *http.Request) {
    var respn model.Response

  // Ambil token dari header Authorization
	authHeader := r.Header.Get("Authorization")
	tokenLogin := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenLogin == "" {
		respn.Status = "Error: Missing Authorization header"
		respn.Info = "Authorization header is missing or invalid."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

    // Temukan user berdasarkan token di database
    userData, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user", bson.M{"token": tokenLogin})
    if err != nil || userData.PhoneNumber == "" {
        respn.Status = "Error: Tidak Diizinkan"
        respn.Info = "Anda tidak memiliki izin untuk mengakses data ini."
        at.WriteJSON(w, http.StatusForbidden, respn)
        return
    }

    // Decode request body untuk update item
    var item model.Itemlarangan
    if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
        at.WriteJSON(w, http.StatusBadRequest, err.Error())
        return
    }

    // Validasi ID item yang diperlukan
    if item.IDItemIND == "" {
        at.WriteJSON(w, http.StatusBadRequest, "ID Item tidak boleh kosong.")
        return
    }
    if item.BarangTerlarang == "" {
        at.WriteJSON(w, http.StatusBadRequest, "Barang Terlarang tidak boleh kosong.")
        return
    }

    // Buat filter berdasarkan id_itemind
    filter := bson.M{"id_itemind": item.IDItemIND}

    // Gunakan $set untuk memperbarui field yang diperlukan
    update := bson.D{
        {Key: "$set", Value: bson.D{
            {Key: "destinasi", Value: item.Destinasi},
            {Key: "barang_terlarang", Value: item.BarangTerlarang},
        }},
    }

    // Update item di database
    if _, err := config.Mongoconn.Collection("prohibited_items_id").UpdateOne(context.TODO(), filter, update); err != nil {
        at.WriteJSON(w, http.StatusInternalServerError, err.Error())
        return
    }

    // Berhasil mengupdate item
    at.WriteJSON(w, http.StatusOK, "Item berhasil diperbarui.")
}

func DeleteitemIND(w http.ResponseWriter, r *http.Request) {
    var respn model.Response

    // Ambil token dari header Authorization
	authHeader := r.Header.Get("Authorization")
	tokenLogin := strings.TrimPrefix(authHeader, "Bearer")
	if tokenLogin == "" {
		respn.Status = "Error: Missing Authorization header"
		respn.Info = "Authorization header is missing or invalid."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

    // Temukan user berdasarkan token di database
    userData, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user", bson.M{"token": tokenLogin})
    if err != nil || userData.PhoneNumber == "" {
        respn.Status = "Error: Tidak Diizinkan"
        respn.Info = "Anda tidak memiliki izin untuk mengakses data ini."
        at.WriteJSON(w, http.StatusForbidden, respn)
        return
    }

    // Decode request body untuk mengambil id_itemind
    var requestBody struct {
        IDItemIND string `json:"id_itemind"`
    }
    if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
        at.WriteJSON(w, http.StatusBadRequest, err.Error())
        return
    }

    if requestBody.IDItemIND == "" {
        respn.Status = "Error: ID Item Hilang"
        respn.Info = "ID Item diperlukan untuk menghapus item."
        at.WriteJSON(w, http.StatusBadRequest, respn)
        return
    }

    // Hapus item berdasarkan id_itemind
    filter := bson.M{"id_itemind": requestBody.IDItemIND}
    result, err := atdb.DeleteOneDoc(config.Mongoconn, "prohibited_items_id", filter)
    if err != nil {
        respn.Status = "Error: Kesalahan Server Internal"
        respn.Info = "Kesalahan saat menghapus item dari database."
        at.WriteJSON(w, http.StatusInternalServerError, respn)
        return
    }

    if result.DeletedCount == 0 {
        respn.Status = "Error: Item Tidak Ditemukan"
        respn.Info = fmt.Sprintf("Tidak ada item yang ditemukan dengan IDItemIND '%s'", requestBody.IDItemIND)
        at.WriteJSON(w, http.StatusNotFound, respn)
        return
    }

    respn.Status = "Sukses"
    respn.Info = fmt.Sprintf("Item dengan IDItemIND '%s' berhasil dihapus.", requestBody.IDItemIND)
    at.WriteJSON(w, http.StatusOK, respn)
}






