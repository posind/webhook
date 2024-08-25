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

func EnsureItemIDExists(w http.ResponseWriter, r *http.Request) error {
	// Cari semua dokumen yang belum memiliki ID atau yang memiliki ID duplikat
	filter := bson.M{
		"$or": []bson.M{
			{"id_itemind": bson.M{"$exists": false}},
			{"id_itemind": ""},
		},
	}
	cursor, err := config.Mongoconn.Collection("prohibited_items_id").Find(context.Background(), filter)
	if err != nil {
		return fmt.Errorf("failed to fetch items: %v", err)
	}
	defer cursor.Close(context.Background())

	var bulkWrites []mongo.WriteModel

	for cursor.Next(context.Background()) {
		var item model.Itemlarangan
		err := cursor.Decode(&item)
		if err != nil {
			return fmt.Errorf("failed to decode item: %v", err)
		}

		// Cek dan buat ID unik untuk item
		isUnique := false
		itemCount := 1
		for !isUnique {
			potentialID := fmt.Sprintf("%s-%03d", item.Destinasi, itemCount)
			existingItem, err := atdb.GetOneDoc[model.Itemlarangan](config.Mongoconn, "prohibited_items_id", bson.M{
				"destinasi":  item.Destinasi,
				"id_itemind": potentialID,
			})
			if err != nil && err != mongo.ErrNoDocuments {
				return fmt.Errorf("failed to check for existing ID: %v", err)
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
	}

	// Eksekusi batch update jika ada dokumen yang perlu diperbarui
	if len(bulkWrites) > 0 {
		_, err := config.Mongoconn.Collection("prohibited_items_id").BulkWrite(context.Background(), bulkWrites)
		if err != nil {
			return fmt.Errorf("failed to execute bulk write: %v", err)
		}
	}

	return nil
}

// GetProhibitedItemByField fetches prohibited items based on the specified destination and prohibited item filters.
func GetitemIND(w http.ResponseWriter, r *http.Request) {
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

func PostitemIND(w http.ResponseWriter, r *http.Request) {
	var newItem model.Itemlarangan
	if err := json.NewDecoder(r.Body).Decode(&newItem); err != nil {
		at.WriteJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	if newItem.Destinasi == "" || newItem.BarangTerlarang == "" {
		at.WriteJSON(w, http.StatusBadRequest, "Destination and Prohibited Items cannot be empty")
		return
	}

	// Pastikan ID item unik untuk semua item di koleksi
	if err := EnsureItemIDExists(w, r); err != nil {
		at.WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Ambil kode destinasi
	destinationCode, err := atdb.GetOneDoc[model.DestinationCode](config.Mongoconn, "destination_code", bson.M{"destination": newItem.Destinasi})
	if err != nil || destinationCode.DestinationID == "" {
		at.WriteJSON(w, http.StatusBadRequest, "Invalid destination")
		return
	}

	// Hitung jumlah item yang ada untuk destinasi tertentu
	itemCount, err := atdb.CountDocs(config.Mongoconn, "prohibited_items_id", bson.M{"destinasi": newItem.Destinasi})
	if err != nil {
		at.WriteJSON(w, http.StatusInternalServerError, "Failed to count existing items for the destination")
		return
	}

	// Buat ID baru berdasarkan kode destinasi dan jumlah item
	newItem.IDItemIND = fmt.Sprintf("%s-%03d", destinationCode.DestinationID, itemCount+1)

	// Simpan item baru ke database
	if _, err := atdb.InsertOneDoc(config.Mongoconn, "prohibited_items_id", newItem); err != nil {
		at.WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	at.WriteJSON(w, http.StatusOK, newItem)
}

func UpdateitemIND(w http.ResponseWriter, r *http.Request) {
	var item model.Itemlarangan
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		at.WriteJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// Pastikan ID item unik untuk semua item di koleksi
	if err := EnsureItemIDExists(w, r); err != nil {
		at.WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Update item berdasarkan IDItemIND
	filter := bson.M{"id_itemind": item.IDItemIND}
	update := bson.M{
		"$set": bson.M{
			"destinasi":        item.Destinasi,
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
