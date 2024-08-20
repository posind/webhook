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
	"github.com/gocroot/helper/passwordhash"
	"github.com/gocroot/model"
)

// GetProhibitedItemByField fetches prohibited items based on the specified destination and prohibited item filters.
func GetitemIND(w http.ResponseWriter, r *http.Request) {
	var respn model.Response

	tokenLogin := r.Header.Get("Login")
	if tokenLogin == "" {
		respn.Status = "Error: Missing Login header"
		respn.Info = "Login header is missing."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	userData, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user_email", bson.M{"token": tokenLogin})
	if err != nil || userData.Email == "" {
		respn.Status = "Error: Unauthorized"
		respn.Info = "You do not have permission to access this data."
		at.WriteJSON(w, http.StatusForbidden, respn)
		return
	}

	decodedUsername, err := passwordhash.DecodeGetUser(userData.Public, tokenLogin)
	if err != nil {
		respn.Status = "Error: Invalid token"
		respn.Info = "The provided token is not valid."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	userByUsername, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user_email", bson.M{"username": decodedUsername})
	if err != nil || userByUsername.Username == "" {
		respn.Status = "Error: User not found"
		respn.Info = fmt.Sprintf("The username '%s' extracted from the token does not exist in the database.", decodedUsername)
		at.WriteJSON(w, http.StatusForbidden, respn)
		return
	}

	query := r.URL.Query()
	destination := query.Get("destinasi")
	prohibitedItems := query.Get("barang_terlarang")

	filterItems := bson.M{}
	if destination != "" {
		filterItems["destinasi"] = destination
	}
	if prohibitedItems != "" {
		filterItems["Barang Terlarang"] = prohibitedItems
	}

	findOptions := options.Find().SetLimit(20)

	var items []model.Itemlarangan
	collection := config.Mongoconn.Collection("prohibited_items_id")

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

	at.WriteJSON(w, http.StatusOK, items)
}

// PostProhibitedItem adds a new prohibited item to the database.
func PostitemIND(w http.ResponseWriter, r *http.Request) {
	var respn model.Response

	tokenLogin := r.Header.Get("Login")
	if tokenLogin == "" {
		respn.Status = "Error: Missing Login header"
		respn.Info = "Login header is missing."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	userData, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user_email", bson.M{"token": tokenLogin})
	if err != nil || userData.Email == "" {
		respn.Status = "Error: Unauthorized"
		respn.Info = "You do not have permission to access this data."
		at.WriteJSON(w, http.StatusForbidden, respn)
		return
	}

	decodedUsername, err := passwordhash.DecodeGetUser(userData.Public, tokenLogin)
	if err != nil {
		respn.Status = "Error: Invalid token"
		respn.Info = "The provided token is not valid."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	userByUsername, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user_email", bson.M{"username": decodedUsername})
	if err != nil || userByUsername.Username == "" {
		respn.Status = "Error: User not found"
		respn.Info = fmt.Sprintf("The username '%s' extracted from the token does not exist in the database.", decodedUsername)
		at.WriteJSON(w, http.StatusForbidden, respn)
		return
	}

	var newItem model.Itemlarangan
	if err := json.NewDecoder(r.Body).Decode(&newItem); err != nil {
		at.WriteJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	if newItem.Destinasi == "" || newItem.BarangTerlarang == "" {
		at.WriteJSON(w, http.StatusBadRequest, "Destination and Prohibited Items cannot be empty")
		return
	}

	var destinationCode model.DestinationCode

	destinationCode, err = atdb.GetOneDoc[model.DestinationCode](config.Mongoconn, "destination_code", bson.M{"destination": newItem.Destinasi})
	if err != nil || destinationCode.DestinationID == "" {
		respn.Status = "Error: Invalid destination"
		respn.Info = "Could not find the country code for the given destination."
		at.WriteJSON(w, http.StatusBadRequest, respn)
		return
	}

	itemCount, err := atdb.CountDocs(config.Mongoconn, "prohibited_items_id", bson.M{"destinasi": newItem.Destinasi})
	if err != nil {
		respn.Status = "Error: Could not generate item ID"
		respn.Info = "Failed to count existing items for the given destination."
		at.WriteJSON(w, http.StatusInternalServerError, respn)
		return
	}

	newItem.IDItemIND = fmt.Sprintf("%s-%03d", destinationCode.DestinationID, itemCount+1)

	if _, err := atdb.InsertOneDoc(config.Mongoconn, "prohibited_items_id", newItem); err != nil {
		at.WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	at.WriteJSON(w, http.StatusOK, newItem)
}

// UpdateProhibitedItem updates an existing prohibited item in the database.
func UpdateitemIND(w http.ResponseWriter, r *http.Request) {
	var respn model.Response

	tokenLogin := r.Header.Get("Login")
	if tokenLogin == "" {
		respn.Status = "Error: Missing Login header"
		respn.Info = "Login header is missing."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	userData, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user_email", bson.M{"token": tokenLogin})
	if err != nil || userData.Email == "" {
		respn.Status = "Error: Unauthorized"
		respn.Info = "You do not have permission to access this data."
		at.WriteJSON(w, http.StatusForbidden, respn)
		return
	}

	decodedUsername, err := passwordhash.DecodeGetUser(userData.Public, tokenLogin)
	if err != nil {
		respn.Status = "Error: Invalid token"
		respn.Info = "The provided token is not valid."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	userByUsername, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user_email", bson.M{"username": decodedUsername})
	if err != nil || userByUsername.Username == "" {
		respn.Status = "Error: User not found"
		respn.Info = fmt.Sprintf("The username '%s' extracted from the token does not exist in the database.", decodedUsername)
		at.WriteJSON(w, http.StatusForbidden, respn)
		return
	}

	var item model.Itemlarangan
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		at.WriteJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	filter := bson.M{"id_itemind": item.IDItemIND}
	update := bson.M{
		"$set": bson.M{
			"destinasi":         item.Destinasi,
			"Barang Terlarang": item.BarangTerlarang,
		},
	}

	if _, err := atdb.UpdateOneDoc(config.Mongoconn, "prohibited_items_id", filter, update); err != nil {
		at.WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	at.WriteJSON(w, http.StatusOK, item)
}

// DeleteProhibitedItem deletes a prohibited item from the database.
func DeleteitemIND(w http.ResponseWriter, r *http.Request) {
	var respn model.Response

	tokenLogin := r.Header.Get("Login")
	if tokenLogin == "" {
		respn.Status = "Error: Missing Login header"
		respn.Info = "Login header is missing."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	userData, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user_email", bson.M{"token": tokenLogin})
	if err != nil || userData.Email == "" {
		respn.Status = "Error: Unauthorized"
		respn.Info = "You do not have permission to access this data."
		at.WriteJSON(w, http.StatusForbidden, respn)
		return
	}

	decodedUsername, err := passwordhash.DecodeGetUser(userData.Public, tokenLogin)
	if err != nil {
		respn.Status = "Error: Invalid token"
		respn.Info = "The provided token is not valid."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	userByUsername, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user_email", bson.M{"username": decodedUsername})
	if err != nil || userByUsername.Username == "" {
		respn.Status = "Error: User not found"
		respn.Info = fmt.Sprintf("The username '%s' extracted from the token does not exist in the database.", decodedUsername)
		at.WriteJSON(w, http.StatusForbidden, respn)
		return
	}

	var item model.Itemlarangan
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		at.WriteJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	filter := bson.M{"id_itemind": item.IDItemIND}

	// Cek apakah item dengan ID tersebut ada
	existingItem, err := atdb.GetOneDoc[model.Itemlarangan](config.Mongoconn, "prohibited_items_id", filter)
	if err != nil || existingItem.IDItemIND == "" {
		respn.Status = "Error: Item not found"
		respn.Info = fmt.Sprintf("No prohibited item found with ID: %s", item.IDItemIND)
		at.WriteJSON(w, http.StatusNotFound, respn)
		return
	}

	// Hapus item
	if _, err := atdb.DeleteOneDoc(config.Mongoconn, "prohibited_items_id", filter); err != nil {
		at.WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	respn.Status = "Success: Item deleted"
	respn.Info = fmt.Sprintf("Prohibited item with ID: %s has been successfully deleted.", item.IDItemIND)
	at.WriteJSON(w, http.StatusOK, respn)
}

// EnsureItemIDExists ensures all items have unique IDItemIND.
func EnsureItemIDExists(w http.ResponseWriter, r *http.Request) {
	var respn model.Response

	tokenLogin := r.Header.Get("Login")
	if tokenLogin == "" {
		respn.Status = "Error: Missing Login header"
		respn.Info = "Login header is missing."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	userData, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user_email", bson.M{"token": tokenLogin})
	if err != nil || userData.Email == "" {
		respn.Status = "Error: Unauthorized"
		respn.Info = "You do not have permission to access this data."
		at.WriteJSON(w, http.StatusForbidden, respn)
		return
	}

	decodedUsername, err := passwordhash.DecodeGetUser(userData.Public, tokenLogin)
	if err != nil {
		respn.Status = "Error: Invalid token"
		respn.Info = "The provided token is not valid."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	userByUsername, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user_email", bson.M{"username": decodedUsername})
	if err != nil || userByUsername.Username == "" {
		respn.Status = "Error: User not found"
		respn.Info = fmt.Sprintf("The username '%s' extracted from the token does not exist in the database.", decodedUsername)
		at.WriteJSON(w, http.StatusForbidden, respn)
		return
	}

	// Temukan semua dokumen yang belum memiliki ID atau yang memiliki ID duplikat
	cursor, err := atdb.FindDocs(config.Mongoconn, "prohibited_items_id", bson.M{
		"$or": []bson.M{
			{"id_itemind": bson.M{"$exists": false}},
			{"id_itemind": ""},
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
		var item model.Itemlarangan
		err := cursor.Decode(&item)
		if err != nil {
			at.WriteJSON(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Cek apakah IDItemIND sudah ada di destinasi yang sama dan buat ID unik
		isUnique := false
		itemCount := 1
		for !isUnique {
			potentialID := fmt.Sprintf("%s-%03d", item.Destinasi, itemCount)
			existingItem, err := atdb.GetOneDoc[model.Itemlarangan](config.Mongoconn, "prohibited_items_id", bson.M{
				"destinasi":   item.Destinasi,
				"id_itemind":  potentialID,
			})
			if err != nil && err != mongo.ErrNoDocuments {
				at.WriteJSON(w, http.StatusInternalServerError, err.Error())
				return
			}
			if existingItem.IDItemIND == "" {
				item.IDItemIND = potentialID
				isUnique = true
			} else {
				itemCount++
			}
		}

		// Siapkan model update untuk bulk write
		updateQuery := bson.M{
			"$set": bson.M{
				"id_itemind": item.IDItemIND,
			},
		}
		filter := bson.M{"_id": item.ID}
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
			bulkWrites = []mongo.WriteModel{}
			counter = 0
		}
	}

	// Berikan respon sukses setelah batch pertama dan berhenti
	if len(bulkWrites) > 0 {
		_, err := config.Mongoconn.Collection("prohibited_items_id").BulkWrite(context.Background(), bulkWrites)
		if err != nil {
			at.WriteJSON(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	at.WriteJSON(w, http.StatusOK, "Prohibited items updated successfully with new IDs where applicable.")
}

