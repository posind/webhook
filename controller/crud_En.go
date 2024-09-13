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

func GetProhibitedItem(w http.ResponseWriter, r *http.Request) {
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

	// Tambahkan log untuk debug
	log.Println("Authorization header received:", authHeader)

	// Find user by token in the database
	userData, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user", bson.M{"token": tokenLogin})
	if err != nil || userData.PhoneNumber == "" {
		respn.Status = "Error: Unauthorized"
		respn.Info = "You do not have permission to access this data."
		at.WriteJSON(w, http.StatusForbidden, respn)
		return
	}

	// Decode the token using the user's public key
	decodedPhoneNumber, err := passwordhash.DecodeGetUser(userData.Public, tokenLogin)
	if err != nil {
		respn.Status = "Error: Invalid token"
		respn.Info = "The provided token is not valid."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	// Check if the decoded phone number exists in the database
	userByPhoneNumber, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user", bson.M{"phonenumber": decodedPhoneNumber})
	if err != nil || userByPhoneNumber.PhoneNumber == "" {
		respn.Status = "Error: User not found"
		respn.Info = fmt.Sprintf("The phone number '%s' extracted from the token does not exist in the database.", decodedPhoneNumber)
		at.WriteJSON(w, http.StatusForbidden, respn)
		return
	}

	// Token is valid and matches an existing user, proceed with fetching the data
	query := r.URL.Query()
	destination := query.Get("destination")
	prohibitedItems := query.Get("prohibited_items")

	filterItems := bson.M{}
	if destination != "" {
		filterItems["destination"] = destination
	}
	if prohibitedItems != "" {
		filterItems["Prohibited Items"] = prohibitedItems
	}

	findOptions := options.Find().SetLimit(20)

	// Query MongoDB
	var items []model.ProhibitedItems
	collection := config.Mongoconn.Collection("prohibited_items_en")

	cursor, err := collection.Find(context.Background(), filterItems, findOptions)
	if err != nil {
		respn.Status = "Error: Internal Server Error"
		respn.Info = "Error fetching items from the database."
		at.WriteJSON(w, http.StatusInternalServerError, respn)
		return
	}
	defer cursor.Close(context.Background())

	if err = cursor.All(context.Background(), &items); err != nil {
		respn.Status = "Error: Internal Server Error"
		respn.Info = "Error decoding items."
		at.WriteJSON(w, http.StatusInternalServerError, respn)
		return
	}

	if len(items) == 0 {
		respn.Status = "Error: No items found"
		respn.Info = "No items match the provided filters."
		at.WriteJSON(w, http.StatusNotFound, respn)
		return
	}

	// Respond with the items as JSON
	at.WriteJSON(w, http.StatusOK, items)
}



func PostProhibitedItem(w http.ResponseWriter, r *http.Request) {
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

	// Tambahkan log untuk debug
	log.Println("Authorization header received:", authHeader)

	// Temukan user berdasarkan token di database
	userData, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user", bson.M{"token": tokenLogin})
	if err != nil || userData.PhoneNumber == "" {
		respn.Status = "Error: Unauthorized"
		respn.Info = "Anda tidak memiliki izin untuk mengakses data ini."
		at.WriteJSON(w, http.StatusForbidden, respn)
		return
	}

	// Decode request body untuk item baru
	var newItem model.ProhibitedItems
	if err := json.NewDecoder(r.Body).Decode(&newItem); err != nil {
		at.WriteJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// Validasi input yang diperlukan
	if newItem.Destination == "" {
		at.WriteJSON(w, http.StatusBadRequest, "Destinasi tidak boleh kosong.")
		return
	}
	if newItem.ProhibitedItems == "" {
		at.WriteJSON(w, http.StatusBadRequest, "Barang Terlarang tidak boleh kosong.")
		return
	}

	// Buat tiga digit acak untuk ID Item
	randomDigits := fmt.Sprintf("%03d", rand.Intn(1000))

	// Cari kode destinasi di database
	var destinationCode model.DestinationCode
	destinationCode, err = atdb.GetOneDoc[model.DestinationCode](config.Mongoconn, "destination_code", bson.M{"destination": newItem.Destination})
	if err != nil || destinationCode.DestinationID == "" {
		respn.Status = "Error: Invalid Destination"
		respn.Info = "Kode negara untuk destinasi yang diberikan tidak ditemukan."
		at.WriteJSON(w, http.StatusBadRequest, respn)
		return
	}

	// Buat ID item otomatis berdasarkan kode negara dan tiga digit acak
	newItem.IDItem = fmt.Sprintf("%s-%s", destinationCode.DestinationID, randomDigits)

	// Masukkan data item ke database
	collection := config.Mongoconn.Collection("prohibited_items_id")
	_, err = collection.InsertOne(context.TODO(), newItem)
	if err != nil {
		at.WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Berhasil menambahkan item
	at.WriteJSON(w, http.StatusOK, newItem)
}



func EnsureIDItemExists(w http.ResponseWriter, r *http.Request) {
	// Temukan semua dokumen yang belum memiliki id_item atau yang memiliki id_item duplikat
	cursor, err := atdb.FindDocs(config.Mongoconn, "prohibited_items_en", bson.M{
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
		var newItem model.ProhibitedItems
		err := cursor.Decode(&newItem)
		if err != nil {
			at.WriteJSON(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Cari kode negara berdasarkan destinasi
		var destinationCode model.DestinationCode
		destinationCode, err = atdb.GetOneDoc[model.DestinationCode](config.Mongoconn, "destination_code", bson.M{"destination": newItem.Destination})
		if err != nil || destinationCode.DestinationID == "" {
			at.WriteJSON(w, http.StatusBadRequest, "Error: Could not find the country code for the given destination.")
			return
		}

		// Periksa apakah id_item sudah ada di destination yang sama dan buat ID yang unik
		isUnique := false
		itemCount := 1
		for !isUnique {
			potentialID := fmt.Sprintf("%s-%03d", destinationCode.DestinationID, itemCount)
			existingItem, err := atdb.GetOneDoc[model.ProhibitedItems](config.Mongoconn, "prohibited_items_en", bson.M{
				"destination": newItem.Destination,
				"id_item":     potentialID,
			})
			if err != nil && err != mongo.ErrNoDocuments {
				at.WriteJSON(w, http.StatusInternalServerError, err.Error())
				return
			}
			if existingItem.IDItem == "" {
				newItem.IDItem = potentialID
				isUnique = true
			} else {
				itemCount++
			}
		}

		// Siapkan update model untuk bulk write, menggunakan filter berdasarkan _id untuk update yang tepat
		updateQuery := bson.M{
			"$set": bson.M{
				"id_item": newItem.IDItem,
			},
		}
		filter := bson.M{"_id": newItem.ID}
		update := mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(updateQuery)
		bulkWrites = append(bulkWrites, update)
		counter++

		// Eksekusi batch jika batchSize tercapai
		if counter >= batchSize {
			_, err := config.Mongoconn.Collection("prohibited_items_en").BulkWrite(context.Background(), bulkWrites)
			if err != nil {
				at.WriteJSON(w, http.StatusInternalServerError, err.Error())
				return
			}
			// Reset batch setelah eksekusi
			at.WriteJSON(w, http.StatusOK, "Prohibited items updated successfully with new IDs where applicable.")
			return
		}
	}

	// Berikan respon sukses setelah batch pertama dan berhenti
	at.WriteJSON(w, http.StatusOK, "Prohibited items updated successfully with new IDs where applicable.")
}

func UpdateProhibitedItem(w http.ResponseWriter, r *http.Request) {
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

	// Tambahkan log untuk debug
	log.Println("Authorization header received:", authHeader)

	// Temukan user berdasarkan token di database
	userData, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user", bson.M{"token": tokenLogin})
	if err != nil || userData.PhoneNumber == "" {
		respn.Status = "Error: Unauthorized"
		respn.Info = "Anda tidak memiliki izin untuk mengakses data ini."
		at.WriteJSON(w, http.StatusForbidden, respn)
		return
	}

	// Decode the token using the user's public key
	decodedPhoneNumber, err := passwordhash.DecodeGetUser(userData.Public, tokenLogin)
	if err != nil {
		respn.Status = "Error: Invalid token"
		respn.Info = "The provided token is not valid."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	// Check if the decoded phone number exists in the database
	userByPhoneNumber, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user", bson.M{"phonenumber": decodedPhoneNumber})
	if err != nil || userByPhoneNumber.PhoneNumber == "" {
		respn.Status = "Error: User not found"
		respn.Info = fmt.Sprintf("The phone number '%s' extracted from the token does not exist in the database.", decodedPhoneNumber)
		at.WriteJSON(w, http.StatusForbidden, respn)
		return
	}

	// Lanjutkan dengan logika asli untuk mengupdate item
	var item model.ProhibitedItems
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		at.WriteJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	filter := bson.M{"id_item": item.IDItem}
	update := bson.M{
		"$set": bson.M{
			"Prohibited Items": item.ProhibitedItems,
		},
	}

	if _, err := atdb.UpdateOneDoc(config.Mongoconn, "prohibited_items_en", filter, update); err != nil {
		at.WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	at.WriteJSON(w, http.StatusOK, item)
}


func DeleteProhibitedItem(w http.ResponseWriter, r *http.Request) {
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

	// Tambahkan log untuk debug
	log.Println("Authorization header received:", authHeader)

	// Temukan user berdasarkan token di database
	userData, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user", bson.M{"token": tokenLogin})
	if err != nil || userData.PhoneNumber == "" {
		respn.Status = "Error: Unauthorized"
		respn.Info = "Anda tidak memiliki izin untuk mengakses data ini."
		at.WriteJSON(w, http.StatusForbidden, respn)
		return
	}

	// Decode the token using the user's public key
	decodedPhoneNumber, err := passwordhash.DecodeGetUser(userData.Public, tokenLogin)
	if err != nil {
		respn.Status = "Error: Invalid token"
		respn.Info = "The provided token is not valid."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	// Check if the decoded phone number exists in the database
	userByPhoneNumber, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user", bson.M{"phonenumber": decodedPhoneNumber})
	if err != nil || userByPhoneNumber.PhoneNumber == "" {
		respn.Status = "Error: User not found"
		respn.Info = fmt.Sprintf("The phone number '%s' extracted from the token does not exist in the database.", decodedPhoneNumber)
		at.WriteJSON(w, http.StatusForbidden, respn)
		return
	}

	// Lanjutkan dengan logika asli untuk menghapus item
	var requestBody struct {
		IDItem string `json:"id_item"`
	}

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		at.WriteJSON(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	// Validasi apakah id_item ada dalam body
	if requestBody.IDItem == "" {
		at.WriteJSON(w, http.StatusBadRequest, "id_item is required")
		return
	}

	// Gunakan atdb.DeleteOneDoc untuk menghapus item dari koleksi "prohibited_items_en"
	_, delErr := atdb.DeleteOneDoc(config.Mongoconn, "prohibited_items_en", bson.M{"id_item": requestBody.IDItem})
	if delErr != nil {
		log.Printf("Error deleting item: %v", delErr)
		at.WriteJSON(w, http.StatusInternalServerError, "Error deleting item")
		return
	}

	at.WriteJSON(w, http.StatusOK, "Item deleted successfully")
}

