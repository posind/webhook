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


func GetItems(w http.ResponseWriter, r *http.Request) {
	collectionEn := config.Mongoconn.Collection("prohibited_items_en")
	collectionId := config.Mongoconn.Collection("prohibited_items_id")

	cursorEn, err := collectionEn.Find(context.Background(), bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursorEn.Close(context.Background())

	cursorId, err := collectionId.Find(context.Background(), bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursorId.Close(context.Background())

	var itemsEn []model.ProhibitedItem_en
	var itemsId []model.ProhibitedItem_id

	if err := cursorEn.All(context.Background(), &itemsEn); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := cursorId.All(context.Background(), &itemsId); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var combinedItems []interface{}
	for _, item := range itemsEn {
		combinedItems = append(combinedItems, item)
	}
	for _, item := range itemsId {
		combinedItems = append(combinedItems, item)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(combinedItems)
}


func CreateItem(w http.ResponseWriter, r *http.Request) {
	var itemEn model.ProhibitedItem_en
	var itemId model.ProhibitedItem_id

	if err := json.NewDecoder(r.Body).Decode(&itemEn); err != nil {
		if err := json.NewDecoder(r.Body).Decode(&itemId); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	if itemEn.Destination != "" {
		collection := config.Mongoconn.Collection("prohibited_items_en")
		_, err := collection.InsertOne(context.Background(), itemEn)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(itemEn)
	} else {
		collection := config.Mongoconn.Collection("prohibited_items_id")
		_, err := collection.InsertOne(context.Background(), itemId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(itemId)
	}
}

func UpdateItem(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var itemEn model.ProhibitedItem_en
	var itemId model.ProhibitedItem_id

	if err := json.NewDecoder(r.Body).Decode(&itemEn); err != nil {
		if err := json.NewDecoder(r.Body).Decode(&itemId); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	filter := bson.M{"_id": objID}
	var update bson.M

	if itemEn.Destination != "" {
		collection := config.Mongoconn.Collection("prohibited_items_en")
		update = bson.M{"$set": itemEn}
		_, err := collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(itemEn)
	} else {
		collection := config.Mongoconn.Collection("prohibited_items_id")
		update = bson.M{"$set": itemId}
		_, err := collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(itemId)
	}
}

func DeleteItem(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	filter := bson.M{"_id": objID}

	collectionEn := config.Mongoconn.Collection("prohibited_items_en")
	collectionId := config.Mongoconn.Collection("prohibited_items_id")

	_, err = collectionEn.DeleteOne(context.Background(), filter)
	if err != nil {
		_, err = collectionId.DeleteOne(context.Background(), filter)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
