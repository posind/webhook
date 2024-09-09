package controller

import (
	"context"
	"encoding/json"
	"fmt"
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

		// Insert new user data
		_, err = atdb.InsertOneDoc(config.Mongoconn, "user", userdata)
		if err != nil {
			resp.Message = "Failed to save user data: " + err.Error()
		} else {
			resp.Status = true
			resp.Message = "User registered successfully"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}


func Login(w http.ResponseWriter, r *http.Request) {
	var resp model.Credential
	var loginReq model.LoginRequest

	// Parsing JSON request dari body
	err := json.NewDecoder(r.Body).Decode(&loginReq)
	if err != nil {
		resp.Message = "Error parsing application/json: " + err.Error()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Mendapatkan data user berdasarkan nomor telepon
	userData, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user", bson.M{"phonenumber": loginReq.PhoneNumber})
	if err != nil || userData.PhoneNumber == "" {
		resp.Message = "User not found"
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Membuat token menggunakan kunci private user
	tokenString, err := watoken.Encode(userData.PhoneNumber, userData.Private)
	if err != nil {
		resp.Message = "Failed to encode token: " + err.Error()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Debug: Cek apakah token dan nomor telepon benar
	fmt.Println("Updating token for phone number:", userData.PhoneNumber)
	fmt.Println("Generated token:", tokenString)

	// Update token in the database using phone number
	update := bson.M{
		"$set": bson.M{
			"token": tokenString,
		},
	}

	_, err = config.Mongoconn.Collection("user").UpdateOne(
		context.Background(),
		bson.M{"phonenumber": userData.PhoneNumber},
		update,
	)
	if err != nil {
		fmt.Println("Error updating token in MongoDB:", err) // Debug log
		resp.Message = "Failed to update token: " + err.Error()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Berikan respon yang sesuai setelah token berhasil diperbarui
	resp.Status = true
	resp.Token = tokenString
	resp.Message = "Login successful"

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}







// func Login(w http.ResponseWriter, r *http.Request) {
// 	var resp model.Credential
// 	var loginReq model.LoginRequest

// 	err := json.NewDecoder(r.Body).Decode(&loginReq)
// 	if err != nil {
// 		resp.Message = "Error parsing application/json: " + err.Error()
// 		w.Header().Set("Content-Type", "application/json")
// 		json.NewEncoder(w).Encode(resp)
// 		return
// 	}

// 	// Get user data based on phone number
// 	userData, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user", bson.M{"phonenumber": loginReq.PhoneNumber})
// 	if err != nil || userData.PhoneNumber == "" {
// 		resp.Message = "User not found"
// 		w.Header().Set("Content-Type", "application/json")
// 		json.NewEncoder(w).Encode(resp)
// 		return
// 	}

// 	// Generate token using user's private key
// 	tokenString, err := watoken.Encode(userData.PhoneNumber, userData.Private)
// 	if err != nil {
// 		resp.Message = "Failed to encode token: " + err.Error()
// 		w.Header().Set("Content-Type", "application/json")
// 		json.NewEncoder(w).Encode(resp)
// 		return
// 	}

// 	// Update token in the database using phone number
// 	update := bson.M{
// 		"$set": bson.M{
// 			"token": tokenString,
// 		},
// 	}

// 	_, err = atdb.UpdateOneDoc(config.Mongoconn, "user", bson.M{"phonenumber": userData.PhoneNumber}, update)
// 	if err != nil {
// 		resp.Message = "Failed to update token: " + err.Error()
// 		w.Header().Set("Content-Type", "application/json")
// 		json.NewEncoder(w).Encode(resp)
// 		return
// 	}

// 	resp.Status = true
// 	resp.Token = tokenString
// 	resp.Message = "Login successful"

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(resp)
// }
