package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/at"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/model"
)

// ProhibitedItem (English) Handlers
func GetProhibitedItemByField(w http.ResponseWriter, r *http.Request) {
	var respn model.Response

	// Extract public key from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		respn.Status = "Error: Missing Authorization header"
		respn.Info = "Authorization header is missing."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {

		respn.Status = "Error: Invalid Authorization header format"
		respn.Info = "Authorization header format is invalid."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
		return
	}

	// Extract token from the header
	token := strings.TrimPrefix(authHeader, "Bearer ")
	// Directly compare the token to the public key
	if token != config.PublicKey {
		respn.Status = "Error: Invalid public key"
		respn.Info = "Public key is not valid."
		at.WriteJSON(w, http.StatusUnauthorized, respn)
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

	log.Printf("Filter created: %+v", filter)

	// Set MongoDB query options
	findOptions := options.Find().SetLimit(20)

	// Query MongoDB
	var items []model.ProhibitedItems
	collection := config.Mongoconn.Collection("prohibited_items_en")

	cursor, err := collection.Find(context.Background(), filter, findOptions)
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
	// Respond with the items
	respn.Status = "Success"
	respn.Response = fmt.Sprintf("%v", items)
	at.WriteJSON(w, http.StatusOK, respn)
}

// PostProhibitedItem adds a new item to the database.
func PostProhibitedItem(w http.ResponseWriter, r *http.Request) {
	var newItem model.ProhibitedItems
	if err := json.NewDecoder(r.Body).Decode(&newItem); err != nil {
		at.WriteJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	newItem.ID = primitive.NewObjectID()

	if newItem.Destination == "" || newItem.ProhibitedItems == "" {
		at.WriteJSON(w, http.StatusBadRequest, "Destination and Prohibited Items cannot be empty")
		return
	}

	if _, err := atdb.InsertOneDoc(config.Mongoconn, "prohibited_items_en", newItem); err != nil {
		at.WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	at.WriteJSON(w, http.StatusOK, newItem)
}

// UpdateProhibitedItem updates an item in the database.
func UpdateProhibitedItem(w http.ResponseWriter, r *http.Request) {
	var item model.ProhibitedItems
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		at.WriteJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	filter := bson.M{"_id": item.ID}
	update := bson.M{
		"$set": bson.M{
			"prohibited_items": item.ProhibitedItems,
		},
	}

	if _, err := atdb.UpdateOneDoc(config.Mongoconn, "prohibited_items_en", filter, update); err != nil {
		at.WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	at.WriteJSON(w, http.StatusOK, item)
}

// DeleteProhibitedItemByField deletes an item based on provided fields.
func DeleteProhibitedItemByField(w http.ResponseWriter, r *http.Request) {
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

	log.Printf("Filter created: %+v", filter)

	if len(filter) == 0 {
		log.Println("No query parameters provided, returning all items.")
	}

	collection := config.Mongoconn.Collection("prohibited_items_en")
	deleteResult, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		log.Printf("Error deleting items: %v", err)
		at.WriteJSON(w, http.StatusInternalServerError, "Error deleting items")
		return
	}

	if deleteResult.DeletedCount == 0 {
		at.WriteJSON(w, http.StatusNotFound, "No items found to delete")
		return
	}

	at.WriteJSON(w, http.StatusOK, "Item deleted successfully")
}
