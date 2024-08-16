package controller

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gocroot/config"
	"github.com/gocroot/helper"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/model"
)

// ProhibitedItem (English) Handlers

// GetProhibitedItemByField fetches items based on provided fields.
func GetProhibitedItemByField(w http.ResponseWriter, r *http.Request) {
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

	var items []model.ProhibitedItems
	collection := config.Mongoconn.Collection("prohibited_items_en")

	// Set options to limit the number of documents returned
	findOptions := options.Find()
	findOptions.SetLimit(20) // Change to 10 if you want to limit to 10 items

	cursor, err := collection.Find(context.Background(), filter, findOptions)
	if err != nil {
		helper.WriteJSON(w, http.StatusInternalServerError, "Error fetching items")
		return
	}
	defer cursor.Close(context.Background())

	if err = cursor.All(context.Background(), &items); err != nil {
		helper.WriteJSON(w, http.StatusInternalServerError, "Error decoding items")
		return
	}

	if len(items) == 0 {
		helper.WriteJSON(w, http.StatusNotFound, "No items found")
		return
	}

	helper.WriteJSON(w, http.StatusOK, items)
}

// PostProhibitedItem adds a new item to the database.
func PostProhibitedItem(w http.ResponseWriter, r *http.Request) {
	var newItem model.ProhibitedItems
	if err := json.NewDecoder(r.Body).Decode(&newItem); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	newItem.ID = primitive.NewObjectID()

	if newItem.Destination == "" || newItem.ProhibitedItems == "" {
		helper.WriteJSON(w, http.StatusBadRequest, "Destination and Prohibited Items cannot be empty")
		return
	}

	if _, err := atdb.InsertOneDoc(config.Mongoconn, "prohibited_items_en", newItem); err != nil {
		helper.WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	helper.WriteJSON(w, http.StatusOK, newItem)
}

// UpdateProhibitedItem updates an item in the database.
func UpdateProhibitedItem(w http.ResponseWriter, r *http.Request) {
	var item model.ProhibitedItems
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		helper.WriteJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	filter := bson.M{"_id": item.ID}
	update := bson.M{
		"$set": bson.M{
			"prohibited_items": item.ProhibitedItems,
		},
	}

	if _, err := atdb.UpdateDoc(config.Mongoconn, "prohibited_items_en", filter, update); err != nil {
		helper.WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	helper.WriteJSON(w, http.StatusOK, item)
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
		helper.WriteJSON(w, http.StatusInternalServerError, "Error deleting items")
		return
	}

	if deleteResult.DeletedCount == 0 {
		helper.WriteJSON(w, http.StatusNotFound, "No items found to delete")
		return
	}

	helper.WriteJSON(w, http.StatusOK, "Item deleted successfully")
}
