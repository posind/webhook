package controller

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gocroot/config"
	"github.com/gocroot/helper"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson"
)

func GetDataId(respw http.ResponseWriter, req *http.Request) {
    resp, err := atdb.GetAllDoc[model.ProhibitedItem_en](config.Mongoconn, "prohibited_items_id", bson.M{})
    if err != nil {
        var respn model.Response
        respn.Response = err.Error()
        helper.WriteJSON(respw, http.StatusInternalServerError, respn)
        return
    }
    helper.WriteJSON(respw, http.StatusOK, resp)
}

func CreateItemId(respw http.ResponseWriter, req *http.Request) {
    var item model.ProhibitedItem_en
    err := json.NewDecoder(req.Body).Decode(&item)
    if err != nil {
        var respn model.Response
        respn.Response = err.Error()
        helper.WriteJSON(respw, http.StatusForbidden, respn)
        return
    }

    _, err = atdb.InsertOneDoc(config.Mongoconn, "prohibited_items_id", item)
    if err != nil {
        var respn model.Response
        respn.Response = err.Error()
        helper.WriteJSON(respw, http.StatusForbidden, respn)
        return
    }

    items, err := atdb.GetAllDoc[model.ProhibitedItem_en](config.Mongoconn, "prohibited_items_id", bson.M{})
    if err != nil {
        var respn model.Response
        respn.Response = err.Error()
        helper.WriteJSON(respw, http.StatusForbidden, respn)
        return
    }

    helper.WriteJSON(respw, http.StatusOK, items)
}

func UpdateItemId(respw http.ResponseWriter, req *http.Request) {
    var item model.ProhibitedItem_en
    err := json.NewDecoder(req.Body).Decode(&item)
    if err != nil {
        var respn model.Response
        respn.Response = err.Error()
        helper.WriteJSON(respw, http.StatusForbidden, respn)
        return
    }

    dt, err := atdb.GetOneDoc[model.ProhibitedItem_en](config.Mongoconn, "prohibited_items_id", bson.M{"_id": item.ID})
    if err != nil {
        var respn model.Response
        respn.Response = err.Error()
        helper.WriteJSON(respw, http.StatusForbidden, respn)
        return
    }

    dt.Destination = item.Destination
    dt.ProhibitedItems = item.ProhibitedItems

    _, err = atdb.ReplaceOneDoc(config.Mongoconn, "prohibited_items_id", bson.M{"_id": item.ID}, dt)
    if err != nil {
        var respn model.Response
        respn.Response = err.Error()
        helper.WriteJSON(respw, http.StatusForbidden, respn)
        return
    }

    helper.WriteJSON(respw, http.StatusOK, dt)
}

func DeleteItemId(respw http.ResponseWriter, req *http.Request) {
    var item model.ProhibitedItem_en
    err := json.NewDecoder(req.Body).Decode(&item)
    if err != nil {
        var respn model.Response
        respn.Response = err.Error()
        helper.WriteJSON(respw, http.StatusForbidden, respn)
        return
    }

    collection := config.Mongoconn.Collection("prohibited_items_id")
    filter := bson.M{"_id": item.ID}

    _, err = collection.DeleteOne(context.Background(), filter)
    if err != nil {
        var respn model.Response
        respn.Response = err.Error()
        helper.WriteJSON(respw, http.StatusInternalServerError, respn)
        return
    }

    helper.WriteJSON(respw, http.StatusOK, map[string]string{"message": "Item deleted successfully"})
}

