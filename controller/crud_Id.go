package controller

import (
	"encoding/json"
	"net/http"

	"github.com/gocroot/config"
	"github.com/gocroot/helper"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/model"
	"go.mongodb.org/mongo-driver/bson"
)

// ProhibitedItem (Indonesian) Handlers

func GetDataProhibitedItemId(respw http.ResponseWriter, req *http.Request) {
	resp, err := atdb.GetAllDoc[[]model.ProhibitedItem_id](config.Mongoconn, "prohibited_items_id", bson.M{})
	if err != nil {
		helper.WriteJSON(respw, http.StatusInternalServerError, model.Response{Response: err.Error()})
		return
	}
	helper.WriteJSON(respw, http.StatusOK, resp)
}

func CreateProhibitedItemId(respw http.ResponseWriter, req *http.Request) {
	var item model.ProhibitedItem_id
	err := json.NewDecoder(req.Body).Decode(&item)
	if err != nil {
		helper.WriteJSON(respw, http.StatusBadRequest, model.Response{Response: err.Error()})
		return
	}

	_, err = atdb.InsertOneDoc(config.Mongoconn, "prohibited_items_id", item)
	if err != nil {
		helper.WriteJSON(respw, http.StatusInternalServerError, model.Response{Response: err.Error()})
		return
	}

	items, err := atdb.GetAllDoc[[]model.ProhibitedItem_id](config.Mongoconn, "prohibited_items_id", bson.M{})
	if err != nil {
		helper.WriteJSON(respw, http.StatusInternalServerError, model.Response{Response: err.Error()})
		return
	}
	helper.WriteJSON(respw, http.StatusOK, items)
}

func UpdateProhibitedItemId(respw http.ResponseWriter, req *http.Request) {
	var item model.ProhibitedItem_id
	err := json.NewDecoder(req.Body).Decode(&item)
	if err != nil {
		helper.WriteJSON(respw, http.StatusBadRequest, model.Response{Response: err.Error()})
		return
	}

	existingItem, err := atdb.GetOneDoc[model.ProhibitedItem_id](config.Mongoconn, "prohibited_items_id", bson.M{"_id": item.ID})
	if err != nil {
		helper.WriteJSON(respw, http.StatusNotFound, model.Response{Response: err.Error()})
		return
	}

	existingItem.Destinasi = item.Destinasi
	existingItem.BarangTerlarang = item.BarangTerlarang

	_, err = atdb.ReplaceOneDoc(config.Mongoconn, "prohibited_items_id", bson.M{"_id": item.ID}, existingItem)
	if err != nil {
		helper.WriteJSON(respw, http.StatusInternalServerError, model.Response{Response: err.Error()})
		return
	}
	helper.WriteJSON(respw, http.StatusOK, existingItem)
}

func DeleteProhibitedItemId(respw http.ResponseWriter, req *http.Request) {
	var item model.ProhibitedItem_id
	err := json.NewDecoder(req.Body).Decode(&item)
	if err != nil {
		helper.WriteJSON(respw, http.StatusBadRequest, model.Response{Response: err.Error()})
		return
	}

    _, err = atdb.DeleteOneDoc(config.Mongoconn, "prohibited_items", bson.M{"_id": item.ID})	
    if err != nil {
		helper.WriteJSON(respw, http.StatusInternalServerError, model.Response{Response: err.Error()})
		return
	}

	items, err := atdb.GetAllDoc[[]model.ProhibitedItem_id](config.Mongoconn, "prohibited_items_id", bson.M{})
	if err != nil {
		helper.WriteJSON(respw, http.StatusInternalServerError, model.Response{Response: err.Error()})
		return
	}
	helper.WriteJSON(respw, http.StatusOK, items)
}

