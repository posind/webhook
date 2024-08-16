package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
		respn.Info = "The username extracted from the token does not exist in the database."
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
		filterItems["prohibited_items"] = prohibitedItems
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
