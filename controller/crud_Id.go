package controller

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gocroot/config"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetItemsId(w http.ResponseWriter, r *http.Request) {
	collection := config.Mongoconn.Collection("prohibited_items_id")
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var items []model.ProhibitedItem_id
	if err := cursor.All(context.Background(), &items); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"/ind": items})
}

func CreateItemId(w http.ResponseWriter, r *http.Request) {
	var item model.ProhibitedItem_id
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	collection := config.Mongoconn.Collection("prohibited_items_id")
	_, err := collection.InsertOne(context.Background(), item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"item_id": item})
}

func UpdateItemId(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var item model.ProhibitedItem_id
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	collection := config.Mongoconn.Collection("prohibited_items_id")
	filter := bson.M{"_id": objID}
	update := bson.M{"$set": item}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"item_id": item})
}

func DeleteItemId(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	collection := config.Mongoconn.Collection("prohibited_items_id")
	filter := bson.M{"_id": objID}

	_, err = collection.DeleteOne(context.Background(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

