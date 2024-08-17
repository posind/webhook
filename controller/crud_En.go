package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/at"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/passwordhash"
	"github.com/gocroot/model"
)

// ProhibitedItem (English) Handlers
func GetProhibitedItemByField(w http.ResponseWriter, r *http.Request) {
	var respn model.Response

	// Extract token from Login header
	tokenLogin := r.Header.Get("Login")
	if tokenLogin == "" {
		respn.Status = "Error: Missing Login header"
		respn.Info = "Login header is missing."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	// Find user by token in the database
	userData, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user_email", bson.M{"token": tokenLogin})
	if err != nil || userData.Email == "" {
		respn.Status = "Error: Unauthorized"
		respn.Info = "You do not have permission to access this data."
		at.WriteJSON(w, http.StatusForbidden, respn)
		return
	}

	// Decode the token using the user's public key
	decodedUsername, err := passwordhash.DecodeGetUser(userData.Public, tokenLogin)
	if err != nil {
		respn.Status = "Error: Invalid token"
		respn.Info = "The provided token is not valid."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	// Check if the decoded username exists in the database
	userByUsername, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user_email", bson.M{"username": decodedUsername})
	if err != nil || userByUsername.Username == "" {
		respn.Status = "Error: User not found"
		respn.Info = fmt.Sprintf("The username '%s' extracted from the token does not exist in the database.", decodedUsername)
		at.WriteJSON(w, http.StatusForbidden, respn)
		return
	}

	// Token is valid and matches an existing user, proceed with fetching the data
	query := r.URL.Query()
	destination := query.Get("destination")
	prohibitedItems := query.Get("prohibited_items")

	log.Printf("Received query parameters - destination: %s, prohibited_items: %s", destination, prohibitedItems)

	filterItems := bson.M{}
	if destination != "" {
		filterItems["destination"] = destination
	}
	if prohibitedItems != "" {
		filterItems["Prohibited Items"] = prohibitedItems
	}

	log.Printf("Filter created: %+v", filterItems)

	// Set MongoDB query options
	findOptions := options.Find().SetLimit(20)

	// Query MongoDB
	var items []model.ProhibitedItems
	collection := config.Mongoconn.Collection("prohibited_items_en")

	cursor, err := collection.Find(context.Background(), filterItems, findOptions)
	if err != nil {
		log.Printf("Error fetching items: %v", err)
		respn.Status = "Error: Internal Server Error"
		respn.Info = "Error fetching items from the database."
		at.WriteJSON(w, http.StatusInternalServerError, respn)
		return
	}
	defer cursor.Close(context.Background())

	if err = cursor.All(context.Background(), &items); err != nil {
		log.Printf("Error decoding items: %v", err)
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

	// Langsung merespons dengan items sebagai JSON
	at.WriteJSON(w, http.StatusOK, items)
}

func PostProhibitedItem(w http.ResponseWriter, r *http.Request) {
	var respn model.Response

	// Ekstrak token dari header Login
	tokenLogin := r.Header.Get("Login")
	if tokenLogin == "" {
		respn.Status = "Error: Missing Login header"
		respn.Info = "Login header is missing."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	// Cari user berdasarkan token di database
	userData, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user_email", bson.M{"token": tokenLogin})
	if err != nil || userData.Email == "" {
		respn.Status = "Error: Unauthorized"
		respn.Info = "You do not have permission to access this data."
		at.WriteJSON(w, http.StatusForbidden, respn)
		return
	}

	// Decode token menggunakan public key user
	decodedUsername, err := passwordhash.DecodeGetUser(userData.Public, tokenLogin)
	if err != nil {
		respn.Status = "Error: Invalid token"
		respn.Info = "The provided token is not valid."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	// Periksa apakah username yang didecode ada di database
	userByUsername, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user_email", bson.M{"username": decodedUsername})
	if err != nil || userByUsername.Username == "" {
		respn.Status = "Error: User not found"
		respn.Info = fmt.Sprintf("The username '%s' extracted from the token does not exist in the database.", decodedUsername)
		at.WriteJSON(w, http.StatusForbidden, respn)
		return
	}

	// Lanjutkan dengan logika asli
	var newItem model.ProhibitedItems
	if err := json.NewDecoder(r.Body).Decode(&newItem); err != nil {
		at.WriteJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	if newItem.Destination == "" || newItem.ProhibitedItems == "" {
		at.WriteJSON(w, http.StatusBadRequest, "Destination and Prohibited Items cannot be empty")
		return
	}

	var destinationCode model.DestinationCode

	destinationCode, err = atdb.GetOneDoc[model.DestinationCode](config.Mongoconn, "destination_code", bson.M{"destination": newItem.Destination})
	if err != nil || destinationCode.DestinationID == "" {
		respn.Status = "Error: Invalid destination"
		respn.Info = "Could not find the country code for the given destination."
		at.WriteJSON(w, http.StatusBadRequest, respn)
		return
	}

	// Hitung jumlah item saat ini untuk negara tersebut untuk membuat ID baru
	itemCount, err := atdb.CountDocs(config.Mongoconn, "prohibited_items_en", bson.M{"destination": newItem.Destination})
	if err != nil {
		respn.Status = "Error: Could not generate item ID"
		respn.Info = "Failed to count existing items for the given destination."
		at.WriteJSON(w, http.StatusInternalServerError, respn)
		return
	}

	// Buat id_item otomatis berdasarkan kode negara dan urutan
	newItem.IDItem = fmt.Sprintf("%s-%03d", destinationCode.DestinationID, itemCount+1)

	// Masukkan data baru ke database
	if _, err := atdb.InsertOneDoc(config.Mongoconn, "prohibited_items_en", newItem); err != nil {
		at.WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	at.WriteJSON(w, http.StatusOK, newItem)
}

func EnsureIDItemExists(w http.ResponseWriter, r *http.Request) {
	// Temukan semua dokumen yang belum memiliki id_item
	cursor, err := atdb.FindDocs(config.Mongoconn, "prohibited_items_en", bson.M{"id_item": bson.M{"$exists": false}})
	if err != nil {
		at.WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer cursor.Close(context.Background())

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

		// Hitung jumlah item yang ada untuk destinasi tersebut
		itemCount, err := atdb.CountDocs(config.Mongoconn, "prohibited_items_en", bson.M{"destination": newItem.Destination})
		if err != nil {
			at.WriteJSON(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Buat id_item otomatis berdasarkan kode negara dan urutan
		newItem.IDItem = fmt.Sprintf("%s-%03d", destinationCode.DestinationID, itemCount+1)

		// Perbarui dokumen dengan id_item baru
		updateQuery := bson.M{
			"$set": bson.M{
				"id_item": newItem.IDItem,
			},
		}
		_, err = atdb.UpdateOneDoc(config.Mongoconn, "prohibited_items_en", bson.M{"id_item": newItem.IDItem}, updateQuery)
		if err != nil {
			at.WriteJSON(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	// Berikan respon sukses setelah semua dokumen diperbarui
	at.WriteJSON(w, http.StatusOK, "Prohibited items updated successfully with new IDs where applicable.")
}

// UpdateProhibitedItem updates an item in the database.
func UpdateProhibitedItem(w http.ResponseWriter, r *http.Request) {
	var respn model.Response

	// Ekstrak token dari header Login
	tokenLogin := r.Header.Get("Login")
	if tokenLogin == "" {
		respn.Status = "Error: Missing Login header"
		respn.Info = "Login header is missing."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	// Cari user berdasarkan token di database
	userData, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user_email", bson.M{"token": tokenLogin})
	if err != nil || userData.Email == "" {
		respn.Status = "Error: Unauthorized"
		respn.Info = "You do not have permission to access this data."
		at.WriteJSON(w, http.StatusForbidden, respn)
		return
	}

	// Decode token menggunakan public key user
	decodedUsername, err := passwordhash.DecodeGetUser(userData.Public, tokenLogin)
	if err != nil {
		respn.Status = "Error: Invalid token"
		respn.Info = "The provided token is not valid."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	// Periksa apakah username yang didecode ada di database
	userByUsername, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user_email", bson.M{"username": decodedUsername})
	if err != nil || userByUsername.Username == "" {
		respn.Status = "Error: User not found"
		respn.Info = fmt.Sprintf("The username '%s' extracted from the token does not exist in the database.", decodedUsername)
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

// DeleteProhibitedItemByField deletes an item based on the provided id_item in the request body.
func DeleteProhibitedItemByField(w http.ResponseWriter, r *http.Request) {
	var respn model.Response

	// Ekstrak token dari header Login
	tokenLogin := r.Header.Get("Login")
	if tokenLogin == "" {
		respn.Status = "Error: Missing Login header"
		respn.Info = "Login header is missing."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	// Cari user berdasarkan token di database
	userData, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user_email", bson.M{"token": tokenLogin})
	if err != nil || userData.Email == "" {
		respn.Status = "Error: Unauthorized"
		respn.Info = "You do not have permission to access this data."
		at.WriteJSON(w, http.StatusForbidden, respn)
		return
	}

	// Decode token menggunakan public key user
	decodedUsername, err := passwordhash.DecodeGetUser(userData.Public, tokenLogin)
	if err != nil {
		respn.Status = "Error: Invalid token"
		respn.Info = "The provided token is not valid."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	// Periksa apakah username yang didecode ada di database
	userByUsername, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user_email", bson.M{"username": decodedUsername})
	if err != nil || userByUsername.Username == "" {
		respn.Status = "Error: User not found"
		respn.Info = fmt.Sprintf("The username '%s' extracted from the token does not exist in the database.", decodedUsername)
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

	filter := bson.M{"id_item": requestBody.IDItem}

	collection := config.Mongoconn.Collection("prohibited_items_en")
	deleteResult, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		log.Printf("Error deleting item: %v", err)
		at.WriteJSON(w, http.StatusInternalServerError, "Error deleting item")
		return
	}

	if deleteResult.DeletedCount == 0 {
		at.WriteJSON(w, http.StatusNotFound, "No item found to delete")
		return
	}

	at.WriteJSON(w, http.StatusOK, "Item deleted successfully")
}
