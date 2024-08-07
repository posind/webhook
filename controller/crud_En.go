package controller

import (
	"encoding/json"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/gocroot/config"
	"github.com/gocroot/helper"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/model"
)

// ProhibitedItem (English) Handlers

func GetDataProhibitedItemEn(respw http.ResponseWriter, req *http.Request) {
	resp, err := atdb.GetAllDoc[[]model.Prohibited_items_en](config.Mongoconn, "prohibited_items_en", bson.M{})
	if err != nil {
		helper.WriteJSON(respw, http.StatusInternalServerError, model.Response{Response: err.Error()})
		return
	}
	helper.WriteJSON(respw, http.StatusOK, resp)
}

func CreateProhibitedItemEn(respw http.ResponseWriter, req *http.Request) {
	var item model.Prohibited_items_en
	err := json.NewDecoder(req.Body).Decode(&item)
	if err != nil {
		helper.WriteJSON(respw, http.StatusBadRequest, model.Response{Response: err.Error()})
		return
	}

	_, err = atdb.InsertOneDoc(config.Mongoconn, "prohibited_items_en", item)
	if err != nil {
		helper.WriteJSON(respw, http.StatusInternalServerError, model.Response{Response: err.Error()})
		return
	}

	items, err := atdb.GetAllDoc[[]model.Prohibited_items_en](config.Mongoconn, "prohibited_items_en", bson.M{})
	if err != nil {
		helper.WriteJSON(respw, http.StatusInternalServerError, model.Response{Response: err.Error()})
		return
	}
	helper.WriteJSON(respw, http.StatusOK, items)
}

func UpdateProhibitedItemEn(respw http.ResponseWriter, req *http.Request) {
	var item model.Prohibited_items_en
	err := json.NewDecoder(req.Body).Decode(&item)
	if err != nil {
		helper.WriteJSON(respw, http.StatusBadRequest, model.Response{Response: err.Error()})
		return
	}

	existingItem, err := atdb.GetOneDoc[model.Prohibited_items_en](config.Mongoconn, "prohibited_items_en", bson.M{"_id": item.ID})
	if err != nil {
		helper.WriteJSON(respw, http.StatusNotFound, model.Response{Response: err.Error()})
		return
	}

	existingItem.Destination = item.Destination
	existingItem.ProhibitedItems = item.ProhibitedItems

	_, err = atdb.ReplaceOneDoc(config.Mongoconn, "prohibited_items_en", bson.M{"_id": item.ID}, existingItem)
	if err != nil {
		helper.WriteJSON(respw, http.StatusInternalServerError, model.Response{Response: err.Error()})
		return
	}
	helper.WriteJSON(respw, http.StatusOK, existingItem)
}

func DeleteProhibitedItemEn(respw http.ResponseWriter, req *http.Request) {
	var item model.Prohibited_items_en
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

	items, err := atdb.GetAllDoc[[]model.Prohibited_items_en](config.Mongoconn, "prohibited_items_en", bson.M{})
	if err != nil {
		helper.WriteJSON(respw, http.StatusInternalServerError, model.Response{Response: err.Error()})
		return
	}
	helper.WriteJSON(respw, http.StatusOK, items)
}