package controller

import (
	"encoding/json"
	"net/http"

	"github.com/gocroot/config"
	"github.com/gocroot/helper/atdb"
	"github.com/gocroot/model"
	"github.com/whatsauth/watoken"
	"go.mongodb.org/mongo-driver/bson"
)

func Register(w http.ResponseWriter, r *http.Request) {
	resp := new(model.Credential)
	userdata := new(model.User)
	resp.Status = false

	// Decode request body to userdata struct
	err := json.NewDecoder(r.Body).Decode(&userdata)
	if err != nil {
		resp.Message = "Error parsing application/json: " + err.Error()
	} else {
		// Check if the phone number is already in use
		existingUser, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user", bson.M{"phonenumber": userdata.PhoneNumber})
		if err == nil && existingUser.PhoneNumber != "" {
			resp.Message = "Phone number already registered"
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		// Generate private and public keys
		privateKey, publicKey := watoken.GenerateKey()
		userdata.Private = privateKey
		userdata.Public = publicKey

		// Insert new user data without email and username
		_, err = atdb.InsertOneDoc(config.Mongoconn, "user", userdata)
		if err != nil {
			resp.Message = "Failed to save user data: " + err.Error()
		} else {
			resp.Status = true
			resp.Message = "User registered successfully"
		}
	}

	// Set response header and return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func Login(w http.ResponseWriter, r *http.Request) {
	var resp model.Credential
	var loginReq model.LoginRequest

	// Decode permintaan login
	err := json.NewDecoder(r.Body).Decode(&loginReq)
	if err != nil {
		resp.Message = "Error parsing application/json: " + err.Error()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest) // Status 400 jika gagal decode
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Ambil data user berdasarkan nomor handphone
	userData, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user", bson.M{"phonenumber": loginReq.PhoneNumber})
	if err != nil || userData.PhoneNumber == "" {
		resp.Message = "User not found!"
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound) // Status 404 jika user tidak ditemukan
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Generate token menggunakan private key
	tokenString, err := watoken.Encode(userData.PhoneNumber, userData.Private)
	if err != nil {
		resp.Message = "Failed to encode the token: " + err.Error()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError) // Status 500 jika gagal encode
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Update token di database menggunakan nomor handphone
	update := bson.M{
		"$set": bson.M{
			"token": tokenString,
		},
	}

	_, err = atdb.UpdateOneDoc(config.Mongoconn, "user", bson.M{"phonenumber": userData.PhoneNumber}, update)
	if err != nil {
		resp.Message = "Failed to update token in database: " + err.Error()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError) // Status 500 jika gagal update
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Set respons sukses setelah token berhasil diperbarui
	resp.Status = true
	resp.Token = tokenString
	resp.Message = "Login successful"
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // Status 200 untuk keberhasilan
	json.NewEncoder(w).Encode(resp)
}



