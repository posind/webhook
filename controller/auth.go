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
		existingUser, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user_email", bson.M{"phonenumber": userdata.PhoneNumber})
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
		_, err = atdb.InsertOneDoc(config.Mongoconn, "user_email", userdata)
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

	err := json.NewDecoder(r.Body).Decode(&loginReq)
	if err != nil {
		resp.Message = "Error parsing application/json: " + err.Error()
		w.Header().Set("Content type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	//mendapatkan user data
	userData, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user_email", bson.M{"phonenumber": loginReq.PhoneNumber})
	if err != nil || userData.PhoneNumber == "" {
		resp.Message = "User not found in database!"
		w.Header().Set("Content type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	// generate token menggunakan private key dari user
	tokenString, err := watoken.Encode(userData.PhoneNumber, userData.Private)
	if err != nil {
		resp.Message = "Failed to encode token: " + err.Error()
		w.Header().Set("Content type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	update := bson.M{
		"$set": bson.M{
			"token": tokenString,
		},
	}

	_, err = atdb.UpdateOneDoc(config.Mongoconn, "user_email", bson.M{"phonenumber": userData.PhoneNumber}, update)
	if err != nil {
		resp.Message = "Failed to update token user in database!"
		w.Header().Set("Content type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp.Status = true
	resp.Token = tokenString
	resp.Message = "Login was succeesful"

	w.Header().Set("Content type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
