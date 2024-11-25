package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/at"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/helper/passwordhash"
	"github.com/gocroot/helper/watoken"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetProhibitedItem(w http.ResponseWriter, r *http.Request) {
	var respn model.Response

	// Extract token from Login header
	tokenLogin := r.Header.Get("login")
	if tokenLogin == "" {
		respn.Status = "Error: Missing Login header"
		respn.Info = "Login header is missing."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	// Decode the token using DecodeGetId to extract the ID (private key in this case)
	privateKey, err := watoken.DecodeGetId(config.PublicKey, tokenLogin)
	if err != nil {
		respn.Status = "Error: Invalid token"
		respn.Info = "The provided token is not valid: " + err.Error()
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	// Find user data using the private key (id) extracted from the token
	userData, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user", bson.M{"private": privateKey})
	if err != nil || userData.PhoneNumber == "" {
		respn.Status = "Error: Unauthorized"
		respn.Info = "You do not have permission to access this data."
		at.WriteJSON(w, http.StatusForbidden, respn)
		return
	}

	// Proceed with fetching prohibited items (your logic for querying items)
	query := r.URL.Query()
	destination := query.Get("destination")
	prohibitedItems := query.Get("prohibited_items")

	// Build filter based on query parameters
	filterItems := bson.M{}
	if destination != "" {
		filterItems["destination"] = destination
	}
	if prohibitedItems != "" {
		filterItems["prohibited_items"] = prohibitedItems
	}

	// Set limit for results
	findOptions := options.Find().SetLimit(20)
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

	// Extract token from Login header
	tokenLogin := r.Header.Get("login")
	if tokenLogin == "" {
		respn.Status = "Error: Missing Login header"
		respn.Info = "Login header is missing."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	// Decode the token to get the user ID
	userID, err := watoken.DecodeGetId(config.PublicKey, tokenLogin)
	if err != nil {
		respn.Status = "Error: Invalid token"
		respn.Info = "The provided token is not valid: " + err.Error()
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	// Find user data using userID
	userData, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user", bson.M{"private": userID})
	if err != nil || userData.PhoneNumber == "" {
		respn.Status = "Error: Unauthorized"
		respn.Info = "You do not have permission to access this data."
		at.WriteJSON(w, http.StatusForbidden, respn)
		return
	}

	// Decode request body for new item
	var newItem model.ProhibitedItems
	if err := json.NewDecoder(r.Body).Decode(&newItem); err != nil {
		at.WriteJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// Validate required input
	if newItem.Destination == "" || newItem.ProhibitedItems == "" || newItem.IDItem == "" {
		at.WriteJSON(w, http.StatusBadRequest, "All fields (Destination, Prohibited Items, ID Item) must be provided.")
		return
	}

	// Insert item into database
	collection := config.Mongoconn.Collection("prohibited_items_en")
	_, err = collection.InsertOne(context.TODO(), bson.M{
		"id_item":         newItem.IDItem,
		"destination":     newItem.Destination,
		"prohibited_items": newItem.ProhibitedItems,
	})
	if err != nil {
		at.WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Respond with the added item
	at.WriteJSON(w, http.StatusOK, bson.M{
		"id_item":         newItem.IDItem,
		"destination":     newItem.Destination,
		"prohibited_items": newItem.ProhibitedItems,
	})
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

	// Extract token from Login header
	tokenLogin := r.Header.Get("login")
	if tokenLogin == "" {
		respn.Status = "Error: Missing Login header"
		respn.Info = "Login header is missing."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	// Decode the token and handle the returned ID and error
	userID, err := watoken.DecodeGetId(config.PublicKey, tokenLogin)
	if err != nil {
		respn.Status = "Error: Invalid token"
		respn.Info = "The provided token is not valid."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	// Find user data using userID
	userData, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user", bson.M{"private": userID})
	if err != nil || userData.PhoneNumber == "" {
		respn.Status = "Error: Unauthorized"
		respn.Info = "You do not have permission to access this data."
		at.WriteJSON(w, http.StatusForbidden, respn)
		return
	}

	// Decode the request body to update the prohibited item
	var item model.ProhibitedItems
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		at.WriteJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// Validate that id_item is present
	if item.IDItem == "" {
		at.WriteJSON(w, http.StatusBadRequest, "ID Item cannot be empty.")
		return
	}

	// Define the filter to find the document by "id_item"
	filter := bson.M{"id_item": item.IDItem}

	// Define the fields to update
	updateFields := bson.M{
		"prohibited_items": item.ProhibitedItems,
		"destination":      item.Destination,
	}

	// Perform the update operation
	result, err := atdb.UpdateOneDoc(config.Mongoconn, "prohibited_items_en", filter, updateFields)
	if err != nil {
		log.Printf("MongoDB Update error: %v", err)
		at.WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	// Check if any documents were modified
	if result.ModifiedCount == 0 {
		respn.Status = "Error: No documents updated"
		respn.Info = "No documents were updated, please check the provided id_item."
		at.WriteJSON(w, http.StatusNotFound, respn)
		return
	}

	// Respond with the updated item
	at.WriteJSON(w, http.StatusOK, bson.M{
		"id_item":          item.IDItem,
		"destination":      item.Destination,
		"prohibited_items": item.ProhibitedItems,
	})
}


func DeleteProhibitedItem(w http.ResponseWriter, r *http.Request) {
	var respn model.Response

	// Extract token from Login header
	tokenLogin := r.Header.Get("login")
	if tokenLogin == "" {
		respn.Status = "Error: Missing Login header"
		respn.Info = "Login header is missing."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	// Decode the token and handle the returned ID and error
	userID, err := watoken.DecodeGetId(config.PublicKey, tokenLogin)
	if err != nil {
		respn.Status = "Error: Invalid token"
		respn.Info = "The provided token is not valid."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	// Find user data using userID
	userData, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user", bson.M{"private": userID})
	if err != nil || userData.PhoneNumber == "" {
		respn.Status = "Error: Unauthorized"
		respn.Info = "You do not have permission to access this data."
		at.WriteJSON(w, http.StatusForbidden, respn)
		return
	}

	// Decode request body
	var requestBody struct {
		IDItem string `json:"id_item"`
	}
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		at.WriteJSON(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	// Validate if id_item is present
	if requestBody.IDItem == "" {
		at.WriteJSON(w, http.StatusBadRequest, "id_item is required")
		return
	}

	// Use atdb.DeleteOneDoc to delete the item
	_, delErr := atdb.DeleteOneDoc(config.Mongoconn, "prohibited_items_en", bson.M{"id_item": requestBody.IDItem})
	if delErr != nil {
		log.Printf("Error deleting item: %v", delErr)
		at.WriteJSON(w, http.StatusInternalServerError, "Error deleting item")
		return
	}

	at.WriteJSON(w, http.StatusOK, "Item deleted successfully")
}

