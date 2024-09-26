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

		// Menghapus field email dan username jika kosong
		userdata.Email = ""      // Hapus email
		userdata.Username = ""   // Hapus username

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

func QRLogin(w http.ResponseWriter, r *http.Request) {
	var resp model.Credential
	var loginReq struct {
		PhoneNumber string `json:"phoneNumber"` // Expecting phoneNumber from QR code
	}

	// Parsing JSON request from body (extracting phoneNumber from QR code)
	err := json.NewDecoder(r.Body).Decode(&loginReq)
	if err != nil {
		sendErrorResponse(w, "Error parsing application/json: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Getting user data based on phoneNumber from QR code
	userData, err := atdb.GetOneDoc[model.User](config.Mongoconn, "user", bson.M{"phonenumber": loginReq.PhoneNumber}) // phonenumber must match the JSON field
	if err != nil {
		// Handle MongoDB errors
		sendErrorResponse(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if user data was found
	if userData.PhoneNumber == "" {
		sendErrorResponse(w, "User not found", http.StatusNotFound)
		return
	}

	// Create token using the found phone number
	tokenString, err := watoken.Encode(userData.PhoneNumber, userData.Private)
	if err != nil {
		sendErrorResponse(w, "Failed to encode token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Debug: Check if token and phone number are correct
	fmt.Println("Updating token for phone number:", userData.PhoneNumber)
	fmt.Println("Generated token:", tokenString)

	// Update token in the database using phoneNumber
	update := bson.M{
		"$set": bson.M{
			"token": tokenString,
		},
	}

	_, err = config.Mongoconn.Collection("user").UpdateOne(
		context.Background(),
		bson.M{"phonenumber": userData.PhoneNumber}, // Match the field correctly
		update,
	)
	if err != nil {
		fmt.Println("Error updating token in MongoDB:", err) // Debug log
		sendErrorResponse(w, "Failed to update token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Add token to Header
	w.Header().Set("Authorization", "Bearer "+tokenString)
	w.Header().Set("Content-Type", "application/json")

	// Provide a response after the token has been successfully updated
	resp.Status = true
	resp.Token = tokenString
	resp.Message = "Login successful via phone number"

	json.NewEncoder(w).Encode(resp)
}




// Helper function to handle error responses
func sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	resp := model.Credential{
		Message: message,
		Status:  false,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
}



// punya teh fahira

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
