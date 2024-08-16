package controller

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gocroot/config"
	"github.com/gocroot/helper"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetItemByField(w http.ResponseWriter, r *http.Request) {
    // Retrieve token from Authorization header
    authHeader := r.Header.Get("Authorization")
    if authHeader == "" {
        log.Println("Authorization header missing")
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    parts := strings.Split(authHeader, " ")
    if len(parts) != 2 || parts[0] != "Bearer" {
        log.Println("Authorization header format invalid")
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    tokenStr := parts[1]
    log.Printf("Received token: %s", tokenStr)

    // Directly parse the public key from environment variable
    publicKeyHex := os.Getenv("PUBLIC_KEY")
    if publicKeyHex == "" {
        log.Println("Public key not found in environment variables")
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    publicKeyBytes, err := hex.DecodeString(publicKeyHex)
    if err != nil {
        log.Printf("Error decoding public key: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    pemBlock := &pem.Block{
        Type:  "PUBLIC KEY",
        Bytes: publicKeyBytes,
    }

    pub, err := x509.ParsePKIXPublicKey(pemBlock.Bytes)
    if err != nil {
        log.Printf("Error parsing public key: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    rsaPublicKey, ok := pub.(*rsa.PublicKey)
    if !ok {
        log.Println("Failed to parse RSA public key")
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // Verify JWT token
    token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
            log.Println("Unexpected signing method")
            return nil, http.ErrAbortHandler
        }
        return rsaPublicKey, nil
    })

    if err != nil || !token.Valid {
        log.Printf("Token invalid or error: %v", err)
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Retrieve query parameters
    query := r.URL.Query()
    destinasi := query.Get("destinasi")
    barang := query.Get("barang")

    // Log query parameters
    log.Printf("Received query parameters - destinasi: %s, barang: %s", destinasi, barang)

    // Build the filter
    filter := bson.M{}
    if destinasi != "" {
        filter["Destinasi"] = destinasi
    }
    if barang != "" {
        filter["Barang"] = barang
    }

    // Log filter
    log.Printf("Filter created: %+v", filter)

    // Set MongoDB query options
    findOptions := options.Find().SetLimit(20)

    // Query MongoDB
    var items []model.Itemlarangan
    collection := config.Mongoconn.Collection("prohibited_items_id")
    cursor, err := collection.Find(context.Background(), filter, findOptions)
    if err != nil {
        log.Printf("Error fetching items: %v", err)
        http.Error(w, "Error fetching items", http.StatusInternalServerError)
        return
    }
    defer cursor.Close(context.Background())

    if err = cursor.All(context.Background(), &items); err != nil {
        log.Printf("Error decoding items: %v", err)
        http.Error(w, "Error decoding items", http.StatusInternalServerError)
        return
    }

    // If no items found
    if len(items) == 0 {
        log.Println("No items found")
        helper.WriteJSON(w, http.StatusNotFound, "No items found")
        return
    }

    // Log found items
    log.Printf("Items found: %+v", items)

    // Return items as JSON
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
    deleteResult, err := collection.DeleteOne(context.Background(), filter)
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


