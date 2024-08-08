package controller

import (
	"encoding/json"
	"net/http"

	"github.com/gocroot/config"
	"github.com/gocroot/helper"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetAllItems mengambil semua item larangan dari database dan mengembalikannya sebagai JSON.
func GetAllItems(respw http.ResponseWriter, req *http.Request) {
	items, err := atdb.GetAllDoc[[]model.Itemlarangan](config.Mongoconn, "item_larangan_id", bson.M{})
	if err != nil {
		helper.WriteJSON(respw, http.StatusInternalServerError, err.Error())
		return
	}
	helper.WriteJSON(respw, http.StatusOK, items)
}

// GetItem mengambil item larangan berdasarkan nama dari database.
func GetItem(respw http.ResponseWriter, req *http.Request) {
	// Ambil parameter dari query string
	nama := req.URL.Query().Get("nama")

	if nama == "" {
		helper.WriteJSON(respw, http.StatusBadRequest, "Missing item name")
		return
	}

	// Buat filter untuk mencari dokumen dengan nama yang diberikan
	filter := bson.M{"barang_terlarang": nama}

	// Ambil satu dokumen item larangan
	item, err := atdb.GetOneDoc[model.Itemlarangan](config.Mongoconn, "item_larangan_id", filter)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			helper.WriteJSON(respw, http.StatusNotFound, "Item not found")
		} else {
			helper.WriteJSON(respw, http.StatusInternalServerError, err.Error())
		}
		return
	}

	// Kembalikan dokumen item larangan dalam format JSON
	helper.WriteJSON(respw, http.StatusOK, item)
}

// CreateItem menambahkan item larangan baru ke dalam database.
func CreateItem(respw http.ResponseWriter, req *http.Request) {
	var newItem model.Itemlarangan
	if err := json.NewDecoder(req.Body).Decode(&newItem); err != nil {
		helper.WriteJSON(respw, http.StatusBadRequest, err.Error())
		return
	}
	newItem.ID = primitive.NewObjectID()

	// Validasi destinasi
	if newItem.Destinasi == "" {
		helper.WriteJSON(respw, http.StatusBadRequest, "Destinasi tidak boleh kosong")
		return
	}

	// Validasi barang terlarang
	if newItem.BarangTerlarang == "" {
		helper.WriteJSON(respw, http.StatusBadRequest, "Barang terlarang tidak boleh kosong")
		return
	}

	if _, err := atdb.InsertOneDoc(config.Mongoconn, "item_larangan_id", newItem); err != nil {
		helper.WriteJSON(respw, http.StatusInternalServerError, err.Error())
		return
	}
	helper.WriteJSON(respw, http.StatusCreated, newItem)
}


// UpdateItem memperbarui item larangan yang ada di database.
func UpdateItem(respw http.ResponseWriter, req *http.Request) {
	var item model.Itemlarangan
	err := json.NewDecoder(req.Body).Decode(&item)
	if err != nil {
		helper.WriteJSON(respw, http.StatusBadRequest, err.Error())
		return
	}

	// Validasi bahwa ID item tidak boleh kosong
	if item.ID == primitive.NilObjectID {
		helper.WriteJSON(respw, http.StatusBadRequest, "ID item tidak boleh kosong")
		return
	}

	// Definisikan filter untuk menemukan item berdasarkan ID item
	filter := bson.M{"_id": item.ID}

	// Definisikan update dengan set data baru
	update := bson.M{
		"$set": bson.M{
			"destinasi":        item.Destinasi,
			"barang_terlarang": item.BarangTerlarang,
		},
	}

	// Update item di MongoDB
	if _, err := atdb.UpdateDoc(config.Mongoconn, "item_larangan_id", filter, update); err != nil {
		helper.WriteJSON(respw, http.StatusInternalServerError, err.Error())
		return
	}

	// Kirim respons sukses
	helper.WriteJSON(respw, http.StatusOK, item)
}


// DeleteItem menghapus item larangan dari database berdasarkan namanya.
func DeleteItem(respw http.ResponseWriter, req *http.Request) {
	var item model.Itemlarangan
	if err := json.NewDecoder(req.Body).Decode(&item); err != nil {
		helper.WriteJSON(respw, http.StatusBadRequest, err.Error())
		return
	}

	// Validasi bahwa nama item tidak boleh kosong
	if item.BarangTerlarang == "" {
		helper.WriteJSON(respw, http.StatusBadRequest, "Nama barang terlarang tidak boleh kosong")
		return
	}

	// Buat filter untuk menghapus item berdasarkan nama barang terlarang
	filter := bson.M{"barang_terlarang": item.BarangTerlarang}
	_, err := atdb.DeleteOneDoc(config.Mongoconn, "item_larangan_id", filter)
	if err != nil {
		helper.WriteJSON(respw, http.StatusInternalServerError, err.Error())
		return
	}
	helper.WriteJSON(respw, http.StatusOK, "Item berhasil dihapus")
}




