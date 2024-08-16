package controller

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"log"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gocroot/config"
	"github.com/gocroot/helper"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetItemByID(respw http.ResponseWriter, req *http.Request) {
    var respn model.Response

    // Log the incoming request headers
    log.Printf("Request headers: %+v", req.Header)

    // Retrieve token from Authorization header
    authHeader := req.Header.Get("Authorization")
    if authHeader == "" {
        respn.Status = "Error: Authorization header missing"
        helper.WriteJSON(respw, http.StatusUnauthorized, respn)
        return
    }

    parts := strings.Split(authHeader, " ")
    if len(parts) != 2 || parts[0] != "Bearer" {
        respn.Status = "Error: Authorization header format invalid"
        helper.WriteJSON(respw, http.StatusUnauthorized, respn)
        return
    }

    tokenStr := parts[1]
    log.Printf("Received token: [REDACTED]")

    // Get the public key from the config package
    publicKeyPEM := config.PublicKey
    if publicKeyPEM == "" {
        log.Println("Public key not found in config")
        respn.Status = "Error: Internal Server Error"
        respn.Info = "Public key missing"
        helper.WriteJSON(respw, http.StatusInternalServerError, respn)
        return
    }

    // Decode PEM block
    block, _ := pem.Decode([]byte(publicKeyPEM))
    if block == nil {
        log.Println("Failed to decode PEM block")
        respn.Status = "Error: Internal Server Error"
        respn.Info = "Failed to decode PEM block"
        helper.WriteJSON(respw, http.StatusInternalServerError, respn)
        return
    }

    // Parse the public key
    pub, err := x509.ParsePKIXPublicKey(block.Bytes)
    if err != nil {
        log.Printf("Error parsing public key: %v", err)
        respn.Status = "Error: Internal Server Error"
        respn.Info = "Failed to parse public key"
        helper.WriteJSON(respw, http.StatusInternalServerError, respn)
        return
    }

    rsaPublicKey, ok := pub.(*rsa.PublicKey)
    if !ok {
        log.Println("Failed to parse RSA public key")
        respn.Status = "Error: Internal Server Error"
        respn.Info = "Public key format is incorrect"
        helper.WriteJSON(respw, http.StatusInternalServerError, respn)
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
        respn.Status = "Error: Unauthorized"
        respn.Info = "Invalid token"
        helper.WriteJSON(respw, http.StatusUnauthorized, respn)
        return
    }

    // Extract the ID from the URL query parameters
    idStr := req.URL.Query().Get("id")
    if idStr == "" {
        respn.Status = "Error: Missing ID parameter"
        helper.WriteJSON(respw, http.StatusBadRequest, respn)
        return
    }

    // Convert the ID string to ObjectID
    objectID, err := primitive.ObjectIDFromHex(idStr)
    if err != nil {
        log.Printf("Invalid ID format: %v", err)
        respn.Status = "Error: Invalid ID format"
        respn.Response = err.Error()
        helper.WriteJSON(respw, http.StatusBadRequest, respn)
        return
    }

    // Log the query
    log.Printf("Querying MongoDB with filter: %+v", bson.M{"_id": objectID})

    // Query MongoDB for the item with the specified ID using GetOneDoc
    item, err := atdb.GetOneDoc[model.Itemlarangan](config.Mongoconn, "prohibited_items_id", bson.M{"_id": objectID})
    if err != nil {
        if err == mongo.ErrNoDocuments {
            log.Println("Item not found")
            respn.Status = "Error: Item not found"
            respn.Info = "No document with the specified ID"
            helper.WriteJSON(respw, http.StatusNotFound, respn)
        } else {
            log.Printf("Error fetching item: %v", err)
            respn.Status = "Error: Internal Server Error"
            respn.Info = "Failed to fetch item"
            helper.WriteJSON(respw, http.StatusInternalServerError, respn)
        }
        return
    }

    // Respond with the item data
    respn.Status = "Success"
    respn.Response = item.BarangTerlarang
    helper.WriteJSON(respw, http.StatusOK, respn)
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


