package controller

import (
	"encoding/json"
	"net/http"

	"github.com/gocroot/config"
	"github.com/gocroot/helper"
	"github.com/gocroot/model"

	"go.mongodb.org/mongo-driver/bson"
)

// GetdataBarang retrieves all prohibited items from the database
func GetdataBarang(respw http.ResponseWriter, req *http.Request) {
    var items []model.Prohibited_items_id
    cursor, err := config.Mongoconn.Collection("prohibited_items_id").Find(req.Context(), bson.D{})
    if err != nil {
        helper.WriteJSON(respw, http.StatusInternalServerError, model.Response{Response: err.Error()})
        return
    }
    defer cursor.Close(req.Context())

    if err = cursor.All(req.Context(), &items); err != nil {
        helper.WriteJSON(respw, http.StatusInternalServerError, model.Response{Response: err.Error()})
        return
    }

    helper.WriteJSON(respw, http.StatusOK, items)
}

// CreateBarang creates a new prohibited item in the database
func CreateBarang(respw http.ResponseWriter, req *http.Request) {
    var item model.Prohibited_items_id
    err := json.NewDecoder(req.Body).Decode(&item)
    if err != nil {
        helper.WriteJSON(respw, http.StatusBadRequest, model.Response{Response: err.Error()})
        return
    }
    _, err = config.Mongoconn.Collection("prohibited_items_id").InsertOne(req.Context(), item)
    if err != nil {
        helper.WriteJSON(respw, http.StatusInternalServerError, model.Response{Response: err.Error()})
        return
    }

    helper.WriteJSON(respw, http.StatusCreated, model.Response{Response: "Item created successfully"})
}

// UpdateBarang updates an existing prohibited item in the database
func UpdateBarang(respw http.ResponseWriter, req *http.Request) {
    var item model.Prohibited_items_id
    err := json.NewDecoder(req.Body).Decode(&item)
    if err != nil {
        helper.WriteJSON(respw, http.StatusBadRequest, model.Response{Response: err.Error()})
        return
    }
    filter := bson.D{{Key: "_id", Value: item.ID}}
    update := bson.D{{Key: "$set", Value: item}}
    _, err = config.Mongoconn.Collection("prohibited_items_id").UpdateOne(req.Context(), filter, update)
    if err != nil {
        helper.WriteJSON(respw, http.StatusInternalServerError, model.Response{Response: err.Error()})
        return
    }

    helper.WriteJSON(respw, http.StatusOK, model.Response{Response: "Item updated successfully"})
}

// DeleteBarang deletes an existing prohibited item from the database
func DeleteBarang(respw http.ResponseWriter, req *http.Request) {
    var item model.Prohibited_items_id
    err := json.NewDecoder(req.Body).Decode(&item)
    if err != nil {
        helper.WriteJSON(respw, http.StatusBadRequest, model.Response{Response: err.Error()})
        return
    }
    filter := bson.D{{Key: "_id", Value: item.ID}}
    _, err = config.Mongoconn.Collection("prohibited_items_id").DeleteOne(req.Context(), filter)
    if err != nil {
        helper.WriteJSON(respw, http.StatusInternalServerError, model.Response{Response: err.Error()})
        return
    }

    helper.WriteJSON(respw, http.StatusOK, model.Response{Response: "Item deleted successfully"})
}


