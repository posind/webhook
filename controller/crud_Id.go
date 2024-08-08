package controller

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gocroot/config"
	"github.com/gocroot/helper"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetItemByField(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query()
    destinasi := query.Get("destinasi")
    barangTerlarang := query.Get("barangTerlarang")

    // Log parameter yang diterima
    log.Printf("Received query parameters - destinasi: %s, barangTerlarang: %s", destinasi, barangTerlarang)

    // Buat filter berdasarkan parameter yang ada
    filter := bson.M{}
    if destinasi != "" {
        filter["destinasi"] = destinasi
    }
    if barangTerlarang != "" {
        filter["barang_terlarang"] = barangTerlarang
    }

    log.Printf("Filter created: %+v", filter)

    // Opsi 1: Jika tidak ada filter, kembalikan semua item
    if len(filter) == 0 {
        log.Println("No query parameters provided, returning all items.")
    }

    // Opsi 2: Jika tidak ada filter, gunakan filter default (ganti sesuai kebutuhan)
    /*
    if len(filter) == 0 {
        filter["destinasi"] = "Default Destination" // Ganti dengan nilai default yang diinginkan
        log.Println("No query parameters provided, using default filter:", filter)
    }
    */

    // Opsi 3: Jika tidak ada filter, kembalikan error
    /*
    if len(filter) == 0 {
        helper.WriteJSON(w, http.StatusBadRequest, "Please provide a valid query parameter")
        return
    }
    */

    // Koneksi ke MongoDB dan gunakan filter untuk mencari dokumen
    var items []model.Itemlarangan
    collection := config.Mongoconn.Collection("prohibited_items_id")
    cursor, err := collection.Find(context.Background(), filter)
    if err != nil {
        helper.WriteJSON(w, http.StatusInternalServerError, "Error fetching items")
        return
    }
    defer cursor.Close(context.Background())

    // Parsing hasil dari cursor MongoDB ke dalam slice item
    if err = cursor.All(context.Background(), &items); err != nil {
        helper.WriteJSON(w, http.StatusInternalServerError, "Error decoding items")
        return
    }

    // Jika tidak ada item yang ditemukan, kirim respons tidak ditemukan
    if len(items) == 0 {
        helper.WriteJSON(w, http.StatusNotFound, "No items found")
        return
    }

    // Kirim hasil item sebagai JSON
    helper.WriteJSON(w, http.StatusOK, items)
}



// PostItem menambahkan item baru ke dalam database.
func PostItem(respw http.ResponseWriter, req *http.Request) {
    var newItem model.Itemlarangan
    if err := json.NewDecoder(req.Body).Decode(&newItem); err != nil {
        helper.WriteJSON(respw, http.StatusBadRequest, err.Error())
        return
    }
    newItem.ID = primitive.NewObjectID()

    // Validasi destinasi dan barang terlarang
    if newItem.Destinasi == "" || newItem.BarangTerlarang == "" {
        helper.WriteJSON(respw, http.StatusBadRequest, "Destinasi dan Barang Terlarang tidak boleh kosong")
        return
    }

    if _, err := atdb.InsertOneDoc(config.Mongoconn, "prohibited_items_id", newItem); err != nil {
        helper.WriteJSON(respw, http.StatusInternalServerError, err.Error())
        return
    }
    helper.WriteJSON(respw, http.StatusOK, newItem)
}

// UpdateItem updates an item in the database.
func UpdateItem(respw http.ResponseWriter, req *http.Request) {
    var item model.Itemlarangan
    if err := json.NewDecoder(req.Body).Decode(&item); err != nil {
        helper.WriteJSON(respw, http.StatusBadRequest, err.Error())
        return
    }

    filter := bson.M{"_id": item.ID}
    update := bson.M{
        "$set": bson.M{
            "barang_terlarang": item.BarangTerlarang,
        },
    }

    // Update item di MongoDB
    if _, err := atdb.UpdateDoc(config.Mongoconn, "prohibited_items_id", filter, update); err != nil {
        helper.WriteJSON(respw, http.StatusInternalServerError, err.Error())
        return
    }

    // Kirim respons sukses
    helper.WriteJSON(respw, http.StatusOK, item)
}

func DeleteItemByField(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query()
    destinasi := query.Get("destinasi")
    barangTerlarang := query.Get("barangTerlarang")

    // Log parameter yang diterima
    log.Printf("Received query parameters - destinasi: %s, barangTerlarang: %s", destinasi, barangTerlarang)

    // Buat filter berdasarkan parameter yang ada
    filter := bson.M{}
    if destinasi != "" {
        filter["destinasi"] = destinasi
    }
    if barangTerlarang != "" {
        filter["barang_terlarang"] = barangTerlarang
    }

    log.Printf("Filter created: %+v", filter)

    // Opsi 1: Jika tidak ada filter, kembalikan semua item
    if len(filter) == 0 {
        log.Println("No query parameters provided, returning all items.")
    }

    // Koneksi ke MongoDB dan gunakan filter untuk menghapus dokumen
    collection := config.Mongoconn.Collection("prohibited_items_id")
    deleteResult, err := collection.DeleteMany(context.Background(), filter)
    if err != nil {
        log.Printf("Error deleting items: %v", err)
        helper.WriteJSON(w, http.StatusInternalServerError, "Error deleting items")
        return
    }

    // Cek apakah ada item yang dihapus
    if deleteResult.DeletedCount == 0 {
        helper.WriteJSON(w, http.StatusNotFound, "No items found to delete")
        return
    }

    // Kirim respon sukses jika item berhasil dihapus
    helper.WriteJSON(w, http.StatusOK, "Items deleted successfully")
}


